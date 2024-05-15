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

	"github.com/lesomnus/tiny-short/bybit"
	"github.com/lesomnus/tiny-short/log"
)

func Run(ctx context.Context, conf *Config) error {
	l, err := conf.Log.NewLogger()
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	l.Info("read config", slog.String("path", conf.path))
	ctx = log.Into(ctx, l)

	actingUser := bybit.AccountInfo{
		Secret: bybit.SecretRecord{
			Type: conf.Secret.Type,
		},
	}
	if data, err := os.ReadFile(conf.Secret.ApiKeyFile); err != nil {
		return fmt.Errorf("read %s: %w", conf.Secret.ApiKeyFile, err)
	} else {
		actingUser.Secret.ApiKey = string(data)
		actingUser.Secret.ApiKey = strings.TrimSpace(actingUser.Secret.ApiKey)
	}
	if data, err := os.ReadFile(conf.Secret.PrivateKeyFile); err != nil {
		return fmt.Errorf("read %s: %w", conf.Secret.PrivateKeyFile, err)
	} else {
		actingUser.Secret.Secret = string(data)
		actingUser.Secret.Secret = strings.TrimSpace(actingUser.Secret.Secret)
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

	client := bybit.NewClient(actingUser.Secret, bybit.WithNetwork(*mainnet))
	if res, err := client.User().QueryApi(ctx, bybit.UserQueryApiReq{}); err != nil {
		return fmt.Errorf("request for user query API: %w", err)
	} else if !res.Ok() {
		return fmt.Errorf("user query API: %w", res.Err())
	} else {
		actingUser.UserId = res.Result.UserId
		actingUser.Secret.DateCreated = res.Result.CreatedAt
		actingUser.Secret.DateExpired = res.Result.ExpiredAt

		h1.Print("API Key Status")
		fmt.Printf(" 🔑\n")
		h2.Print("UID ")
		fmt.Println(actingUser.UserId)
		h2.Print("Created at ")
		fmt.Println(actingUser.Secret.DateCreated)
		h2.Print("Expired at ")
		fmt.Print(actingUser.Secret.DateExpired, " ... ")

		h2.Print("left ")
		h2.Println(DurationString(actingUser.Secret.Expiry()))

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
		if !conf.Move.Enabled {
			fmt.Println("= Disabled")
		} else if slices.Contains(res.Result.Permissions.Wallet, "SubMemberTransfer") {
			p_good.Println("✓ SubMemberTransfer")
		} else {
			isGood = false
			p_fail.Print("✗ SubMemberTransfer ")
			p_fail_why.Println("required to transfer asset between accounts")
		}

		h2.Print("    Main Account ")
		if !conf.Move.Enabled {
			fmt.Println("= Disabled")
		} else if res.Result.IsMaster {
			p_good.Println("✓ OK")
			actingUser.Username = "$MAIN"
		} else {
			isGood = false
			p_fail.Print("✗ Sub Account ")
			p_fail_why.Println("need to be a main account to transfer asset")
		}

		if !isGood {
			return errors.New("check list not satisfied")
		}
	}

	// Assert:
	//   If `conf.Move.Enabled`` == true, `actingUser` must be a main account.
	h1.Println("\nResolve user IDs...")
	if !conf.Move.Enabled {
		fmt.Println("  Skipped as no transfer required")
	} else {
		res, err := client.User().QuerySubMembers(ctx, bybit.UserQuerySubMembersReq{})
		if err != nil {
			return fmt.Errorf("request for query sub members: %w", err)
		} else if !res.Ok() {
			return fmt.Errorf("query sub members: %w", res.Err())
		}

		// Fills user IDs by username.
		conf.Move.from = make([]bybit.AccountInfo, len(conf.Move.From))
		users := make([]*bybit.AccountInfo, len(conf.Move.From)+1)
		users[0] = &conf.Move.to
		users[0].Username = conf.Move.To
		for i := range conf.Move.from {
			users[i+1] = &conf.Move.from[i]
			users[i+1].Username = conf.Move.From[i]
		}

	L:
		for _, u := range users {
			if u.Username == "$MAIN" {
				*u = actingUser
				continue
			}

			for _, v := range res.Result.SubMembers {
				if u.Username == v.Username {
					u.UserId = v.UserId
					continue L
				}
			}

			p_fail.Printf("User %s not found, abort.\n", u.Username)
			return fmt.Errorf("user %s not found", u.Username)
		}

		// Fills secrets.
		errs := make([]error, 0)
		for _, u := range users {
			if u.Username == "$MAIN" {
				h2.Printf("%9s ", u.UserId)
				p_dimmed.Print("$MAIN..............")
				p_good.Println("✓ OK")
				continue
			}

			h2.Printf("%9s ", u.UserId)
			p_dimmed.Printf("%s...", u.Username)
			if u.Username != conf.Move.to.Username {
				// Sub user that does not need API key.
				p_good.Print("✓ OK \n")
				continue
			}

			if s, ok := secrets.Get(u.UserId); ok {
				u.Secret = s
			} else if s, err := createSubApiKey(ctx, client, *u); err != nil {
				h2.Printf("%9s ", u.UserId)
				p_dimmed.Printf("%s ", u.Username)
				p_fail.Print("✗ Create API key ")
				p_fail_why.Printf("%s\n", err.Error())
				errs = append(errs, fmt.Errorf("create sub API key: %w", err))
				continue
			} else {
				u.Secret = s
				secrets.Set(u.UserId, s)
			}

			p_good.Print("✓ OK ")
			fmt.Printf("🔑 left %s\n", DurationString(u.Secret.Expiry()))
		}

		if conf.Secret.Store.Enabled {
			if err := func() error {
				f, err := os.OpenFile(conf.Secret.Store.Path, os.O_RDWR|os.O_CREATE, 0600)
				if err != nil {
					return fmt.Errorf("open secret store: %w", err)
				}
				if err := bybit.SaveSecrets(f, secrets); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				p_fail.Printf("Failed to save secrets at %s ", conf.Secret.Store.Path)
				p_fail_why.Println(err.Error())
				errs = append(errs, fmt.Errorf("save secrets at %s: %w", conf.Secret.Store.Path, err))
			}

		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}
	}

	exec := Exec{
		Client:  client,
		Move:    conf.Move,
		Debug:   conf.Debug,
		Secrets: secrets,
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
