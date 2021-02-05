package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const liveDir = "/etc/letsencrypt/live"

type config struct {
	AcmeServer            string
	AlternativeNames      string
	CertName              string
	CommonName            string
	DNSWaitSeconds        int
	GCSBucket             string
	GCSPrefix             string
	GoogleCredentialsFile string
	RenewalDays           int
}

func main() {
	if err := mainCmd(); err != nil {
		log.Fatalf("certifier: %v", err)
	}
}

func mainCmd() error {
	var cfg config
	flag.StringVar(&cfg.AcmeServer, "acme-server", "https://acme-v02.api.letsencrypt.org/directory", "")
	flag.StringVar(&cfg.AlternativeNames, "alternative-names", "", "")
	flag.StringVar(&cfg.CertName, "cert-name", "", "")
	flag.StringVar(&cfg.CommonName, "common-name", "", "")
	flag.IntVar(&cfg.DNSWaitSeconds, "dns-wait-seconds", 240, "")
	flag.StringVar(&cfg.GCSBucket, "gcs-bucket", "", "")
	flag.StringVar(&cfg.GCSPrefix, "gcs-prefix", "", "")
	flag.IntVar(&cfg.RenewalDays, "renewal-days", 15, "")
	flag.Parse()

	googleCredentialsFile, found := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if !found {
		return errors.New("GOOGLE_APPLICATION_CREDENTIALS not set")
	}
	cfg.GoogleCredentialsFile = googleCredentialsFile

	files := []string{
		filepath.Join(cfg.CertName, "cert.pem"),
		filepath.Join(cfg.CertName, "chain.pem"),
		filepath.Join(cfg.CertName, "fullchain.pem"),
		filepath.Join(cfg.CertName, "privkey.pem"),
	}

	gcs := gcs{
		bucket: cfg.GCSBucket,
		prefix: cfg.GCSPrefix,
	}

	live := live{
		liveDir: liveDir,
	}

	err := gcs.test(context.Background())
	if err != nil {
		log.Printf("Error from GCS test: %v", err)
		return err
	}

	// Find existing certificate (if any) and determine if it needs to be renewed.
	renew := func() bool {
		certPath := filepath.Join(cfg.CertName, "cert.pem")
		body, err := gcs.load(context.Background(), certPath)
		if err != nil {
			log.Printf("Failed to load certificate from GCS")
			return true
		}

		renew, err := shouldRenew(body, cfg.RenewalDays)
		if err != nil {
			log.Printf("Failed to parse certificate data")
			return true
		}

		return renew
	}()

	if !renew {
		log.Printf("Not renewing certificate")
		return nil
	}
	log.Printf("Renewing certificate")

	// Run certbot to recreate our certs
	cmd := buildCertbotCommand(cfg)
	if err := cmd.Run(); err != nil {
		catCertbotLog()
		return err
	}

	catCertbotLog()

	// Copy all generated files to GCS
	var privkey, fullchain []byte
	for _, file := range files {
		log.Printf("Loading %q from disk", file)
		data, err := live.load(file)
		if err != nil {
			return err
		}
		log.Printf("Loaded %q from disk", file)

		log.Printf("Saving %q to GCS", file)
		if err := gcs.save(context.Background(), file, data); err != nil {
			return err
		}
		log.Printf("Saved %q to GCS", file)

		if strings.Contains(file, "privkey.pem") {
			privkey = data
		}
		if strings.Contains(file, "fullchain.pem") {
			fullchain = data
		}
	}

	file := filepath.Join(cfg.CertName, "privkey-plus-fullchain.pem")
	log.Printf("Saving %q to GCS", file)
	if err := gcs.save(context.Background(), file, append(privkey, fullchain...)); err != nil {
		return err
	}
	log.Printf("Saved %q to GCS", file)

	return nil
}

func buildCertbotCommand(cfg config) *exec.Cmd {
	args := []string{
		"certonly",
		"--agree-tos",
		"--break-my-certs",
		"--cert-name", cfg.CertName,
		"--dns-google",
		"--dns-google-credentials", cfg.GoogleCredentialsFile,
		"--dns-google-propagation-seconds", strconv.Itoa(cfg.DNSWaitSeconds),
		"--domain", cfg.CommonName,
		"--email", "infra@stackrox.com",
		"--force-renewal",
		"--non-interactive",
		"--preferred-challenges", "dns",
		"--server", cfg.AcmeServer,
	}

	if len(cfg.AlternativeNames) != 0 {
		args = append(args, "--domains", cfg.AlternativeNames)
	}

	cmd := exec.Command("certbot", args...)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func catCertbotLog() {
	cmd := exec.Command("cat", "/var/log/letsencrypt/letsencrypt.log")
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Println(">>> Certbot Log <<<")
	if err := cmd.Run(); err != nil {
		log.Printf("Error dumping log: %v\n", err)
	}
	log.Println(">>> End Certbot Log <<<")
}

// Check if the current cert is close enough to its expiration date to warrant renewal.
func shouldRenew(certBytes []byte, days int) (bool, error) {
	block, _ := pem.Decode(certBytes)
	if block == nil {
		return false, fmt.Errorf("failed to parse pem file")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, err
	}

	log.Printf("Parsed cert for %+v (expires on %v)\n", cert.Subject, cert.NotAfter)

	timeRemaining := time.Until(cert.NotAfter)
	timeGrace := time.Duration(days*24) * time.Hour

	if timeRemaining <= timeGrace {
		log.Printf("Renewing certificate since time remaining (%v) is less than the grace period (%v)\n", timeRemaining, timeGrace)
		return true, nil
	}

	log.Printf("Not renewing certificate since time remaining (%v) is greater than the grace period (%v)\n", timeRemaining, timeGrace)
	return false, nil
}
