package key

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

var Gen = &cli.Command{
	Name:  "gen",
	Usage: "generates RSA key pair",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "prv-out",
			Value: "prv_key.pem",
			Usage: "output path for a private key",
		},
		&cli.StringFlag{
			Name:  "pub-out",
			Value: "pub_key.pem",
			Usage: "output path for a public key",
		},
	},
	Action: func(c *cli.Context) error {
		prv_out_p := c.String("prv-out")
		pub_out_p := c.String("pub-out")

		prv_out, err := touch(prv_out_p)
		if err != nil {
			return fmt.Errorf("touch %s: %w", prv_out_p, err)
		}
		defer prv_out.Close()

		pub_out, err := touch(pub_out_p)
		if err != nil {
			return fmt.Errorf("touch %s: %w", pub_out_p, err)
		}
		defer pub_out.Close()

		prv_key, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return fmt.Errorf("generate RSA key: %w", err)
		}

		if err := pem.Encode(prv_out, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(prv_key),
		}); err != nil {
			return fmt.Errorf("encode private key into pem: %w", err)
		}

		pub_key_pem := bytes.Buffer{}
		if err := pem.Encode(&pub_key_pem, &pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&prv_key.PublicKey),
		}); err != nil {
			return fmt.Errorf("encode public key into pem: %w", err)
		}
		if _, err := pub_out.Write(pub_key_pem.Bytes()); err != nil {
			return fmt.Errorf("write public key at %s: %w", pub_out_p, err)
		}

		p := color.New(color.FgHiWhite)
		p.Println("key generated:")
		fmt.Print("private key: ")
		p.Println(prv_out_p)
		fmt.Print(" public key: ")
		p.Println(pub_out_p)

		p.Printf("\nCopy your public key to ByBit!:\n")
		fmt.Println(pub_key_pem.String())

		return nil
	},
}
