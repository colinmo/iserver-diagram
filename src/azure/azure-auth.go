package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/browser"
)

type MSAuthResponse struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresDate  time.Time
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AzureAuth struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func (azure *AzureAuth) StartAzure() {
	azure.Init()
	azure.StartLocalServers()
	azure.Login()
}

func (a *AzureAuth) Init() {
}

var AuthWebServer *http.Server
var waitForAuth chan bool

func (a *AzureAuth) Login() {
	waitForAuth = make(chan bool)
	browser.OpenURL(
		fmt.Sprintf(`https://login.microsoftonline.com/%s/oauth2/v2.0/authorize?finalUri=?code=xy&client_id=%s&response_type=code&redirect_uri=http://localhost:10089/auth&response_mode=query&scope=%s`,
			AZURE_TENANT_ID,
			AZURE_CLIENT_ID,
			AZURE_SCOPES),
	)
	for {
		select {
		case <-waitForAuth:
			return
		default:
			time.Sleep(time.Second * 2)
		}
	}
}

func (a *AzureAuth) StartLocalServers() {
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		a.Authenticate(w, r)
	})
	go func() {
		AuthWebServer = &http.Server{Addr: ":10089", Handler: nil}
		if err := AuthWebServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}

func (a *AzureAuth) Authenticate(w http.ResponseWriter, r *http.Request) {
	var AZToken MSAuthResponse
	query := r.URL.Query()
	if query.Get("code") != "" {
		payload := url.Values{
			"client_id":     {AZURE_CLIENT_ID},
			"scope":         {AZURE_SCOPES},
			"code":          {query.Get("code")},
			"redirect_uri":  {"http://localhost:10089/auth"},
			"grant_type":    {"authorization_code"},
			"client_secret": {AZURE_CLIENT_SECRET},
			//"requested_token_use": {"on_behalf_of"},
		}
		resp, err := http.PostForm(
			fmt.Sprintf(
				`https://login.microsoftonline.com/%s/oauth2/v2.0/token`,
				AZURE_TENANT_ID,
			),
			payload,
		)
		if err != nil || resp.StatusCode != http.StatusOK {
			defer resp.Body.Close()
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			fmt.Printf("Body %s\n", bodyString)
			log.Fatalf("Login failed %s\n", err)
		} else {
			err := json.NewDecoder(resp.Body).Decode(&AZToken)
			if err != nil {
				log.Fatalf("Failed MS %s\n", err)
			}
			a.RefreshToken = AZToken.RefreshToken
			seconds, _ := time.ParseDuration(fmt.Sprintf("%ds", AZToken.ExpiresIn-10))
			a.ExpiresAt = time.Now().Add(seconds)
			w.Header().Add("Content-type", "text/html")
			fmt.Fprintf(w, "<html><head></head><body><H1>Authenticated<p>You are authenticated, you may close this window.</body></html>")
			a.AccessToken = AZToken.AccessToken
			if len(a.AccessToken) == 0 {
				log.Fatalf("Could not retrieve access token from the code sent %v", AZToken)
			}
			waitForAuth <- true
		}
	}
}

func (a *AzureAuth) TokenRefresh() {
	var AZToken MSAuthResponse
	if len(a.RefreshToken) == 0 {
		log.Fatal("No refresh token")
		return
	}
	payload := url.Values{
		"client_id":     {AZURE_CLIENT_ID},
		"scope":         {AZURE_SCOPES},
		"refresh_token": {a.RefreshToken},
		"redirect_uri":  {"http://localhost:10089/auth"},
		"grant_type":    {"refresh_token"},
		"client_secret": {AZURE_CLIENT_SECRET},
	}
	resp, err := http.PostForm(
		fmt.Sprintf(`https://login.microsoftonline.com/%s/oauth2/v2.0/token`,
			AZURE_TENANT_ID,
		),
		payload,
	)
	if err != nil {
		log.Fatalf("Login failed %s\n", err)
	} else {
		err := json.NewDecoder(resp.Body).Decode(&AZToken)
		if err != nil {
			log.Fatalf("Failed MS %s\n", err)
		}
		a.RefreshToken = AZToken.RefreshToken
		seconds, _ := time.ParseDuration(fmt.Sprintf("%ds", AZToken.ExpiresIn-10))
		a.ExpiresAt = time.Now().Add(seconds)
		a.AccessToken = AZToken.AccessToken
	}
}

func (a *AzureAuth) CallRestEndpoint(method string, path string, payload []byte, query string) (io.ReadCloser, error) {
	for {
		if len(a.AccessToken) > 0 {
			break
		}
	}
	if a.ExpiresAt.Before(time.Now()) {
		fmt.Printf("Refresh")
		a.TokenRefresh()
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	newpath, _ := url.JoinPath("https://griffith-api.iserver365.com/", path)
	if len(query) > 0 {
		newpath = newpath + "?" + query
	}
	req, _ := http.NewRequest(method, newpath, bytes.NewReader(payload))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.AccessToken))
	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		return resp.Body, err
	}
	if err == nil {
		resultMessage := ""
		if resp.Body != nil {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			resultMessage = string(bodyBytes)
		}
		return resp.Body, fmt.Errorf("iserver query failure, received %d\n%s\n%s", resp.StatusCode, newpath, resultMessage)
	}
	return nil, err
}
