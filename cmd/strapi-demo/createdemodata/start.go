/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package createdemodata

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	cmdutils "github.com/trustbloc/edge-store/pkg/utils/cmd"
)

const (
	adminURLFlagName      = "admin-url"
	adminURLFlagShorthand = "a"
	adminURLFlagUsage     = "URL to run the strapi-demo instance on. Format: HostName:Port."
	adminURLEnvKey        = "STRAPI-DEMO_ADMIN_URL"
	adminURLEndpoint      = "/admin/auth/local/register"
	studentCardsEndpoint  = "/studentcards"
	transcriptEndpoint    = "/transcripts"
)

var errMissingAdminURL = errors.New("admin URL not provided")

type strapiDemoParameters struct {
	client   *http.Client
	adminURL string
}
type strapiUser struct {
	Jwt  string `json:"jwt"`
	User user   `json:"user"`
}
type user struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"isAdmin"`
}

// GetStartCmd returns the Cobra start command.
func GetStartCmd() *cobra.Command {
	startCmd := createStartCmd()

	createFlags(startCmd)

	return startCmd
}
func createStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-demo-data",
		Short: "Create demo data",
		Long:  "Start populating data in strapi with default studentcards and transcripts",
		RunE: func(cmd *cobra.Command, args []string) error {
			hostURL, err := cmdutils.GetUserSetVar(cmd, adminURLFlagName, adminURLEnvKey, false)
			if err != nil {
				return err
			}
			parameters := &strapiDemoParameters{
				client:   &http.Client{},
				adminURL: hostURL,
			}
			return startStrapiDemo(parameters)
		},
	}
}

func createFlags(startCmd *cobra.Command) {
	startCmd.Flags().StringP(adminURLFlagName, adminURLFlagShorthand, "", adminURLFlagUsage)
}

// For Demo you can verify the records by browsing http://localhost:1337/admin/
func startStrapiDemo(parameters *strapiDemoParameters) error {
	if parameters.adminURL == "" {
		return errMissingAdminURL
	}

	var client = parameters.client

	adminUserValues := map[string]string{
		"username": "strapi",
		"email":    "user@strapi.io",
		"password": "strapi"}

	authToken, err := createAdminUser(client, parameters.adminURL, adminUserValues)
	if err != nil {
		return err
	}
	// dummy data for demo purposes
	studentRecord1 := map[string]string{
		"studentid":  "1234568",
		"name":       "Tanu",
		"university": "Faber College",
		"semester":   "3",
		"issuedate":  "2019-01-02T00:00:00.000Z",
	}
	studentRecord2 := map[string]string{
		"studentid":  "323456898",
		"name":       "Derek",
		"university": "Faber College",
		"semester":   "2",
		"issuedate":  "2019-03-02T00:00:00.000Z",
	}
	transcriptRecord1 := map[string]string{
		"studentid":    "323456898",
		"name":         "Tanu",
		"university":   "Faber College",
		"status":       "graduated",
		"totalcredits": "100",
		"course":       "Bachelors'in Computing Science",
	}
	transcriptRecord2 := map[string]string{
		"studentid":    "1234568",
		"name":         "Derek",
		"university":   "Faber College",
		"status":       "graduated",
		"totalcredits": "200",
		"course":       "Bachelors'in Computing Science",
	}

	createOrFetchRecord(client, authToken, parameters.adminURL+adminURLEndpoint, "POST", studentRecord1)
	createOrFetchRecord(client, authToken, parameters.adminURL+studentCardsEndpoint, "POST", studentRecord2)
	createOrFetchRecord(client, authToken, parameters.adminURL+transcriptEndpoint, "POST", transcriptRecord1)
	createOrFetchRecord(client, authToken, parameters.adminURL+transcriptEndpoint, "POST", transcriptRecord2)
	createOrFetchRecord(client, authToken, parameters.adminURL+studentCardsEndpoint, "GET", nil)
	createOrFetchRecord(client, authToken, parameters.adminURL+transcriptEndpoint, "GET", nil)

	return nil
}

// createAdminUser creates the admin user and generates the JWT token
func createAdminUser(client *http.Client, adminURL string, adminUserValues interface{}) (string, error) {
	jsonValue, err := json.Marshal(adminUserValues)

	if err != nil {
		return "", err
	}

	resp, err := client.Post(adminURL+adminURLEndpoint, "application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		return "", err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	body, e := ioutil.ReadAll(resp.Body)

	if e != nil {
		return "", err
	}

	var adminUser strapiUser
	err = json.Unmarshal(body, &adminUser)

	if err != nil {
		return "", err
	}

	token := fmt.Sprintf("%v", adminUser.Jwt)

	return "Bearer " + token, nil
}

// createOrFetchRecord Create the record in CMS and fetch the records too
func createOrFetchRecord(client *http.Client, authToken, url, method string, record interface{}) {
	requestBody, err := json.Marshal(record)
	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		panic(err)
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", authToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	log.Println(string(body))
}
