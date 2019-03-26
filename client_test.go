package transferwise

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// TestNewClient confirms that a client can be created with the default baseURL
// and default User-Agent.
func TestNewClient(t *testing.T) {
	c := NewClient(nil)

	if got, want := c.baseURL.String(), SandboxURL; got != want {
		t.Errorf("client should use default url: got %s, want %s", got, want)
	}

	if got, want := c.userAgent, userAgent; got != want {
		t.Errorf("client should use default user-agent: got %s, want %s", got, want)
	}
}

// TestNewClientWithOptions confirms that a new client can be created by passing
// in custom ClientOptions.
func TestNewClientWithOptions(t *testing.T) {
	baseURL, _ := url.Parse("https://dummyurl:4000")
	userAgent := "dummy user agent"

	opts := ClientOptions{
		baseURL:   baseURL,
		userAgent: userAgent,
	}
	c := NewClientWithOptions(nil, opts)

	if got, want := c.baseURL.String(), baseURL.String(); got != want {
		t.Errorf("client should use custom url: got %s, want %s", got, want)
	}

	if got, want := c.userAgent, userAgent; got != want {
		t.Errorf("client should use custom user-agent: got %s, want %s", got, want)
	}
}

// TestNewRequest confirms that new client requests are created with the
// correct path, userAgent and body.
func TestNewRequest(t *testing.T) {
	client := NewClient(nil)
	method := http.MethodGet
	path := "/test"
	body := struct {
		TestName    string `json:"name"`
		PackageName string `json:"pkg"`
	}{
		TestName:    "TestNewRequest",
		PackageName: "github.com/billglover/transferwise",
	}
	bodyJSON := `{"name":"TestNewRequest","pkg":"github.com/billglover/transferwise"}`

	t.Run("invalid path", func(st *testing.T) {
		_, err := client.NewRequest(method, path+string(byte(0x00)), body)
		if err == nil {
			st.Error("expected an error: got none")
		}
	})

	t.Run("invalid body", func(st *testing.T) {
		body := struct {
			Func func() error
		}{
			Func: func() error { return nil },
		}
		_, err := client.NewRequest(method, path, body)
		if err == nil {
			st.Error("expected an error: got none")
		}
	})

	t.Run("invalid method", func(st *testing.T) {
		_, err := client.NewRequest(string(byte(0x00)), path, body)
		if err == nil {
			st.Error("expected an error: got none")
		}
	})

	t.Run("valid request", func(st *testing.T) {
		req, err := client.NewRequest(method, path, body)

		if err != nil {
			st.Fatalf("unexpected error: got %v", err)
		}

		if got, want := req.URL.String(), SandboxURL+path; got != want {
			st.Errorf("invalid path: got %s, want %s", got, want)
		}

		if got, want := req.UserAgent(), userAgent; got != want {
			st.Errorf("invalid user-agent: got %s, want %s", got, want)
		}

		b, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			st.Fatalf("unexpected error: got %v", err)
		}

		if got, want := strings.TrimSpace(string(b)), bodyJSON; got != want {
			st.Errorf("unexpected body:\n\tgot %s\n\twant %s", got, want)
		}
	})
}

func TestDo(t *testing.T) {

	t.Run("successful POST request", func(st *testing.T) {

		client, mux, _, teardown := setup()
		defer teardown()

		method := http.MethodPost

		type foo struct{ A string }

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if got, want := r.Method, method; got != want {
				st.Errorf("unexpected method: got %s, want %s", got, want)
			}
			fmt.Fprint(w, `{"A":"a"}`)
		})

		want := &foo{"a"}
		got := new(foo)

		req, _ := client.NewRequest(method, "", nil)
		_, err := client.Do(context.Background(), req, got)
		if err != nil {
			st.Fatalf("unexpected error: got %v", err)
		}

		if !reflect.DeepEqual(got, want) {
			st.Error("unexpected response body")
		}
	})

}

// Setup establishes a test Server that can be used to provide mock responses during testing.
// It returns a pointer to a client, a mux, the server URL and a teardown function that
// must be called when testing is complete.
func setup() (*Client, *http.ServeMux, string, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	client := NewClient(nil)
	url, _ := url.Parse(server.URL + "/")
	client.baseURL = url

	return client, mux, server.URL, server.Close
}
