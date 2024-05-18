package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/lesomnus/tiny-short/bybit"
	"github.com/lesomnus/tiny-short/log"
)

func Root(ctx context.Context, conf *Config) error {
	l, err := conf.Log.NewLogger()
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	l.Info("read config", slog.String("path", conf.path))
	ctx = log.Into(ctx, l)

	acting_account := bybit.AccountInfo{
		Secret: bybit.SecretRecord{
			Type: conf.Secret.Type,
		},
	}
	if data, err := os.ReadFile(conf.Secret.ApiKeyFile); err != nil {
		return fmt.Errorf("read %s: %w", conf.Secret.ApiKeyFile, err)
	} else {
		acting_account.Secret.ApiKey = string(data)
		acting_account.Secret.ApiKey = strings.TrimSpace(acting_account.Secret.ApiKey)
	}
	if data, err := os.ReadFile(conf.Secret.PrivateKeyFile); err != nil {
		return fmt.Errorf("read %s: %w", conf.Secret.PrivateKeyFile, err)
	} else {
		acting_account.Secret.Secret = string(data)
		acting_account.Secret.Secret = strings.TrimSpace(acting_account.Secret.Secret)
	}

	secrets := bybit.SecretStore{}
	if conf.Secret.Store.Enabled {
		f, err := os.OpenFile(conf.Secret.Store.Path, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return fmt.Errorf("open secret store: %w", err)
		}
		if err := bybit.LoadSecrets(f, secrets); err != nil {
			return fmt.Errorf("load secrets at %s: %w", conf.Secret.Store.Path, err)
		}
	}

	mainnet, err := url.Parse(bybit.MainNetAddr1)
	if err != nil {
		return fmt.Errorf("invalid main net url: %w", err)
	}

	client := bybit.NewClient(acting_account.Secret, bybit.WithNetwork(*mainnet))
	if res, err := client.User().QueryApi(ctx, bybit.UserQueryApiReq{}); err != nil {
		return fmt.Errorf("request for user query API: %w", err)
	} else if !res.Ok() {
		return fmt.Errorf("user query API: %w", res.Err())
	} else {
		acting_account.UserId = res.Result.UserId
		acting_account.Secret.DateCreated = res.Result.CreatedAt
		acting_account.Secret.DateExpired = res.Result.ExpiredAt

		fmt.Print("🔑 ")
		h1.Print("API Key Status\n")
		h2.Print("UID ")
		fmt.Println(acting_account.UserId)
		h2.Print("Created at ")
		fmt.Println(acting_account.Secret.DateCreated)
		h2.Print("Expired at ")
		fmt.Print(acting_account.Secret.DateExpired, " ... ")

		h2.Print("left ")
		h2.Println(DurationString(time.Until(acting_account.Secret.DateExpired)))

		isGood := true
		h2.Println("Check List:")
		h2.Print("    Read & Write ")
		if res.Result.ReadOnly == 0 {
			p_good.Println("✓ OK")
		} else {
			p_fail.Print("✗ Read Only")
			p_fail_why.Println("need write permission")
		}

		h2.Print("  Contract Trade ")
		if slices.Contains(res.Result.Permissions.ContractTrade, "Order") {
			p_good.Println("✓ Order")
		} else {
			isGood = false
			p_fail.Print("✗ Order ")
			p_fail_why.Println("required to make short order")
		}

		h2.Print("     Derivatives ")
		if slices.Contains(res.Result.Permissions.Derivatives, "DerivativesTrade") {
			p_good.Println("✓ DerivativesTrade")
		} else {
			isGood = false
			p_fail.Print("✗ DerivativesTrade ")
			p_fail_why.Println("required to make short order")
		}

		h2.Print("        Transfer ")
		if !conf.Transfer.Enabled {
			fmt.Println("= Disabled")
		} else if slices.Contains(res.Result.Permissions.Wallet, "SubMemberTransfer") {
			p_good.Println("✓ SubMemberTransfer")
		} else {
			isGood = false
			p_fail.Print("✗ SubMemberTransfer ")
			p_fail_why.Println("required to transfer asset between accounts")
		}

		h2.Print("    Main Account ")
		if !conf.Transfer.Enabled {
			fmt.Println("= Disabled")
		} else if res.Result.IsMaster {
			p_good.Println("✓ OK")
			acting_account.Username = "$MAIN"
		} else {
			isGood = false
			p_fail.Print("✗ Sub Account ")
			p_fail_why.Println("need to be a main account to transfer asset")
		}

		if !isGood {
			return errors.New("check list not satisfied")
		}
	}

	fmt.Println()

	transfer_plan := TransferPlan{}
	if !conf.Transfer.Enabled {
		transfer_plan.Users = []bybit.AccountInfo{acting_account}
	} else {
		transfer_plan.Users = make([]bybit.AccountInfo, len(conf.Transfer.From)+1)
		users := transfer_plan.Users

		fmt.Print("🪪  ")
		h1.Print("Resolve User IDs\n")

		// Assert:
		//   If `conf.Move.Enabled` == true, `actingUser` must be a main account.
		if res, err := client.User().QuerySubMembers(ctx, bybit.UserQuerySubMembersReq{}); err != nil {
			return fmt.Errorf("request for query sub members: %w", err)
		} else if !res.Ok() {
			return fmt.Errorf("query sub members: %w", res.Err())
		} else {
			// Fills user IDs by username.
			users[0].Nickname = conf.Transfer.To.Nickname
			users[0].Username = conf.Transfer.To.Username
			for i, a := range conf.Transfer.From {
				users[i+1].Nickname = a.Nickname
				users[i+1].Username = a.Username
			}

			failed := false
			for i := range users {
				u := &users[i]
				ok := false
				if u.Username == "$MAIN" {
					ok = true
					*u = acting_account
				} else {
					for _, v := range res.Result.SubMembers {
						if u.Username == v.Username {
							ok = true
							u.UserId = v.UserId
							break
						}
					}
				}

				h2.Printf("%8s ", u.DisplayNameTrunc(8))
				p_dimmed.Printf("%s ", u.UserId.String())
				if ok {
					p_good.Printf("✓ OK\n")
				} else {
					p_fail.Printf("✓ Not found\n")
				}

				failed = !ok
			}

			if failed {
				return fmt.Errorf("some users are not found")
			}
		}

		u := &users[0]
		if u.Username != "$Main" {
			fmt.Println()
			h2.Print("Getting trading account's API key... ")

			if s, ok := secrets.Get(u.UserId); ok && time.Until(s.DateExpired) > 96*time.Hour {
				u.Secret = s
				p_good.Print("✓ OK ")
				p_dimmed.Print("from secret store ")
			} else if s, err := createSubApiKey(ctx, client, *u); err != nil {
				p_fail.Print("✗ Failed to create API key ")
				p_fail_why.Printf("%s\n", err.Error())
				return fmt.Errorf("create trading account's API key: %w", err)
			} else {
				u.Secret = s
				secrets.Set(u.UserId, s)

				p_good.Print("✓ OK ")
				p_dimmed.Print("new API key is created ")
			}

			fmt.Printf("🔑 left %s\n", DurationString(time.Until(u.Secret.DateExpired)))

			if conf.Secret.Store.Enabled {
				if err := func() error {
					f, err := os.OpenFile(conf.Secret.Store.Path, os.O_RDWR|os.O_CREATE, 0600)
					if err != nil {
						return fmt.Errorf("open secret store: %w", err)
					}
					defer f.Close()

					if err := bybit.SaveSecrets(f, secrets); err != nil {
						return err
					}

					return nil
				}(); err != nil {
					p_fail.Printf("Failed to save secrets at %s ", conf.Secret.Store.Path)
					p_fail_why.Println(err.Error())
					return fmt.Errorf("save secrets at %s: %w", conf.Secret.Store.Path, err)
				}
			}
		}
	}

	exec := Exec{
		Client:       client,
		TransferPlan: transfer_plan,
		Debug:        conf.Debug,
		Secrets:      secrets,
	}

	errs := make([]error, 0)
	for _, coin := range conf.Coins {
		if err := exec.Do(ctx, coin); err != nil {
			errs = append(errs, fmt.Errorf("execution failed %s: %w", coin, err))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func createSubApiKey(ctx context.Context, client bybit.Client, account bybit.AccountInfo) (bybit.SecretRecord, error) {
	s := bybit.SecretRecord{
		Type: bybit.SecretTypeHmac,
	}

	if res, err := client.User().CreateSubApiKey(ctx, bybit.UserCreateSubApiKeyReq{
		SubUserId: account.UserId,
		Note:      "tiny-short",
		ReadOnly:  0,
		Permissions: bybit.ApiPermissions{
			ContractTrade: []string{"Order"},
		},
	}); err != nil {
		return bybit.SecretRecord{}, fmt.Errorf("request for create sub APi key: %w", err)
	} else if !res.Ok() {
		return bybit.SecretRecord{}, fmt.Errorf("create sub APi key: %w", res.Err())
	} else {
		s.ApiKey = res.Result.ApiKey
		s.Secret = res.Result.Secret
	}

	c := client.Clone(s)
	if res, err := c.User().QueryApi(ctx, bybit.UserQueryApiReq{}); err != nil {
		return bybit.SecretRecord{}, fmt.Errorf("request for user query API: %w", err)
	} else if !res.Ok() {
		return bybit.SecretRecord{}, fmt.Errorf("user query API: %w", res.Err())
	} else {
		s.DateCreated = res.Result.CreatedAt
		s.DateExpired = res.Result.ExpiredAt
	}

	return s, nil
}
