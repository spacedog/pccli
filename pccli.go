package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli"
)

//Constants
const pccliVersion = "0.1.0"
const apiURL = "https://packagecloud.io/api/v1"
const userAgent = "pccli go client"

//Types
type apiClient struct {
	token string
	user  string
}

type pkg struct {
	Name        string `json:"name"`
	VersionsURL string `json:"versions_url"`
}

func newAPIClient(c *cli.Context) (*apiClient, error) {

	t, tsuccess := getArg("apikey", c)
	u, usuccess := getArg("user", c)
	if tsuccess && usuccess {
		return &apiClient{token: t, user: u}, nil
	}
	return nil, errors.New("apikey or user must be set")

}

func (c apiClient) ListPackages(repo string, pkgtype string, distro string) ([]pkg, error) {
	requestURL := fmt.Sprintf("%s/repos/%s/%s/packages/%s/%s.json", apiURL, c.user, repo, pkgtype, distro)
	req, err := http.NewRequest("GET", requestURL, nil)

	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.token, "")
	req.Header.Add("User-Agent", userAgent)

	hc := &http.Client{}
	resp, err := hc.Do(req)

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		var pkgs []pkg
		json.NewDecoder(resp.Body).Decode(&pkgs)
		return pkgs, nil
	default:
		fmt.Println(resp.StatusCode)
		return nil, errors.New("Unexpected HTTP status")
	}

}

//Flags

var flgCommon = []cli.Flag{
	cli.StringFlag{
		Name:  "apikey",
		Usage: "API key to use for authorization",
	},
	cli.StringFlag{
		Name:  "user",
		Usage: "Owner of packagecloud account",
	},
	cli.StringFlag{
		Name:  "config",
		Usage: "Configuration file path",
	},
}

func cmdPackageList() cli.Command {

	flgCommand := []cli.Flag{
		cli.StringFlag{
			Name:  "repo",
			Usage: "repository name",
		},
		cli.StringFlag{
			Name:  "distro",
			Usage: "package distribution",
			Value: "el",
		},
		cli.StringFlag{
			Name:  "pkgtype",
			Usage: "package type",
			Value: "rpm",
		},
	}
	allFlags := append(flgCommon, flgCommand...)
	cmd := cli.Command{
		Name:  "packagelist",
		Flags: allFlags,
		Usage: "List all packages in the repository",
		Action: func(c *cli.Context) error {
			actionListPackage(c)
			return nil
		},
	}
	return cmd
}

// Action function
func actionListPackage(c *cli.Context) {
	ac, err := newAPIClient(c)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	repo, success := getArg("repo", c)
	if !success {
		fmt.Println("repo must be set")
		return
	}
	pkgType, success := getArg("pkgtype", c)
	if !success {
		fmt.Println("pkgtype must be set")
		return
	}
	distro, success := getArg("distro", c)
	if !success {
		fmt.Println("distro must be set")
		return
	}

	pkgs, err := ac.ListPackages(repo, pkgType, distro)

	if err == nil {
		for i, p := range pkgs {
			fmt.Printf("%3d: %s\n", i+1, p.Name)
		}
	}

}

// Helper functions
func getArg(s string, c *cli.Context) (string, bool) {
	if len(c.String(s)) > 0 {
		return c.String(s), true
	}
	return "", false

}

func main() {
	app := cli.NewApp()
	app.Name = "pccli"
	app.Version = pccliVersion
	app.Description = "Packagecloud command line interface"
	app.Commands = []cli.Command{
		cmdPackageList(),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
