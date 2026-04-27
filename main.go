package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"inn2-oauth/oauthclient"
	"io"
	"os"
	"strings"
)

// ReadAuthRequest reads an authentication request from r in the following format:
//
//	ClientAuthname: user\r\n
//	ClientPassword: pass\r\n
//	.\r\n
//
// Fields are parsed case-insensitively. Lines may be terminated by CRLF or LF.
// Returns the client auth name and password, or an error if the input is malformed
// or the terminating dot-line is missing.
func ReadAuthRequest(r io.Reader) (clientAuthname, clientPassword string, err error) {
	scanner := bufio.NewScanner(r)
	var seenAuth, seenPass bool
	for scanner.Scan() {
		line := scanner.Text()
		// scanner.Text() drops the final '\n' but may keep a trailing '\r'.
		if strings.HasSuffix(line, "\r") {
			line = strings.TrimSuffix(line, "\r")
		}
		if line == "." {
			if !seenAuth || !seenPass {
				return "", "", errors.New("incomplete auth request: missing fields")
			}
			return clientAuthname, clientPassword, nil
		}
		if strings.TrimSpace(line) == "" {
			// ignore empty lines
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			// ignore unknown/malformed lines rather than failing immediately
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch strings.ToLower(key) {
		case "clientauthname":
			clientAuthname = value
			seenAuth = true
		case "clientpassword":
			clientPassword = value
			seenPass = true
		default:
			// unknown header; ignore
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", err
	}
	return "", "", errors.New("unexpected EOF: missing terminator")
}

func main() {

	var oauthCfgFile string
	flag.StringVar(&oauthCfgFile, "config", "/etc/news/oauth-login-inn2.yaml", "Path to OAuth config file")
	flag.Parse()

	oauthCfg, err := oauthclient.LoadOauthConfig(oauthCfgFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error loading OAuth config: %v\n", err)
		os.Exit(1)
	}

	// Small CLI wrapper: read auth request from stdin and print parsed values.
	authname, password, err := ReadAuthRequest(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	userParts := strings.Split(authname, "@")
	var domain string
	if len(userParts) == 2 {
		domain = userParts[1]
	} else {
		domain = os.Getenv("DEFAULT_DOMAIN")
		if domain == "" {
			domain = "default"
		}
	}
	client, err := oauthCfg.GetClient(domain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting client for domain %s: %v\n", domain, err)
		os.Exit(3)
	}
	var accessToken string

	if client.OauthTokenUsername != "" && client.OauthTokenUsername == authname {
		accessToken = password
	} else {
		tokenResponse, err := client.ObtainAccessToken(authname, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error obtaining access token: %v\n", err)
			os.Exit(4)
		}
		accessToken = tokenResponse.AccessToken
	}

	userInfo, err := client.CheckIdentity(accessToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error checking identity: %v\n", err)
		os.Exit(5)
	}
	for _, field := range client.UsernameFields {
		if val, ok := userInfo[field]; ok {
			authname = val
			break
		}
	}

	fmt.Printf("User:%s", authname)

	os.Exit(0)
}
