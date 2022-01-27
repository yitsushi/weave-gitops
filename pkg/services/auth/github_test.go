package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	fakehttp "github.com/weaveworks/weave-gitops/pkg/vendorfakes/http"
)

type testServerTransport struct {
	testServeUrl string
	roundTripper http.RoundTripper
}

func (t *testServerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// Fake out the client but preserve the URL, as the URLs are key to validating that
	// the authHandler is working.
	tsUrl, err := url.Parse(t.testServeUrl)
	if err != nil {
		return nil, err
	}

	tsUrl.Path = r.URL.Path

	r.URL = tsUrl

	return t.roundTripper.RoundTrip(r)
}

type outcome struct {
	string
	error
}

type timeSpy struct {
	mutex  sync.Mutex
	time   time.Time
	sleeps []time.Duration
}

func (t *timeSpy) sleep(d time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.time = t.time.Add(d)
	t.sleeps = append(t.sleeps, d)
}

func (t *timeSpy) now() time.Time {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.time
}

var _ = Describe("Github Device Flow", func() {
	var ts *httptest.Server
	var client *http.Client
	token := "gho_sUpErSecRetToKeN"
	userCode := "ABC-123"
	verificationUri := "http://somegithuburl.com"

	var _ = BeforeEach(func() {
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// Quick and dirty router to simulate the Github API
			if strings.Contains(r.URL.Path, "/device/code") {
				err := json.NewEncoder(w).Encode(&GithubDeviceCodeResponse{
					DeviceCode:      "123456789",
					UserCode:        userCode,
					VerificationURI: verificationUri,
					Interval:        1,
				})
				Expect(err).NotTo(HaveOccurred())

			}

			if strings.Contains(r.URL.Path, "/oauth/access_token") {
				err := json.NewEncoder(w).Encode(&githubAuthResponse{
					AccessToken: token,
					Error:       "",
				})
				Expect(err).NotTo(HaveOccurred())
			}
		}))

		client = ts.Client()
		client.Transport = &testServerTransport{testServeUrl: ts.URL, roundTripper: client.Transport}
	})

	var _ = AfterEach(func() {
		ts.Close()
	})

	It("does the auth flow", func() {
		authHandler := NewGithubDeviceFlowHandler(client)

		var cliOutput bytes.Buffer
		result, err := authHandler(context.Background(), &cliOutput)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(token))
		// We need to ensure the user code and verification url are in the CLI ouput.
		// Check for the prescense of substrings to avoid failing tests on trivial output changes.
		Expect(cliOutput.String()).To(ContainSubstring(userCode))
		Expect(cliOutput.String()).To(ContainSubstring(verificationUri))
	})

	FDescribe("pollAuthStatus", func() {
		It("retries after a slow_down response from github", func() {
			t := timeSpy{}
			rt := newMockRoundTripper(1, token, t.now)
			client.Transport = &testServerTransport{testServeUrl: ts.URL, roundTripper: rt}
			interval := 5 * time.Second

			outcomeChan := make(chan outcome)

			go func() {
				resultToken, err := pollAuthStatus(t.sleep, interval, client, "somedevicecode")
				outcomeChan <- outcome{resultToken, err}
			}()

			expectedSleeps := []time.Duration{
				interval,
				interval + 5*time.Second,
			}

			<-outcomeChan
			Expect(t.sleeps).To(Equal(expectedSleeps))
		})

		It("keeps slowing down when told to", func() {
			t := timeSpy{}
			rt := newMockRoundTripper(3, token, t.now)
			client.Transport = &testServerTransport{testServeUrl: ts.URL, roundTripper: rt}
			interval := 5 * time.Second

			outcomeChan := make(chan outcome)

			go func() {
				resultToken, err := pollAuthStatus(t.sleep, interval, client, "somedevicecode")
				outcomeChan <- outcome{resultToken, err}
			}()

			expectedSleeps := []time.Duration{
				interval,
				interval + 5*time.Second,
				interval + 10*time.Second,
				interval + 15*time.Second,
			}

			<-outcomeChan
			Expect(t.sleeps).To(Equal(expectedSleeps))
			pollTimes := []time.Time{}
			for pollTime := range rt.callChan {
				pollTimes = append(pollTimes, pollTime)
			}
			Expect(pollTimes).To(Equal([]time.Time{
				time.Time{}.Add(expectedSleeps[0]),
				time.Time{}.Add(expectedSleeps[0] + expectedSleeps[1]),
				time.Time{}.Add(expectedSleeps[0] + expectedSleeps[1] + expectedSleeps[2]),
				time.Time{}.Add(expectedSleeps[0] + expectedSleeps[1] + expectedSleeps[2] + expectedSleeps[3]),
			}))
		})

		It("returns a token after a slow_down", func() {
			t := timeSpy{}
			rt := newMockRoundTripper(1, token, t.now)
			client.Transport = &testServerTransport{testServeUrl: ts.URL, roundTripper: rt}
			interval := 5 * time.Second

			outcomeChan := make(chan outcome)

			go func() {
				resultToken, err := pollAuthStatus(t.sleep, interval, client, "somedevicecode")
				outcomeChan <- outcome{resultToken, err}
			}()

			outcome := <-outcomeChan

			Expect(outcome.string).To(Equal(token))
			Expect(outcome.error).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("ValidateToken", func() {
	It("returns unauthenticated on an invalid token", func() {
		rt := &fakehttp.FakeRoundTripper{}
		gh := NewGithubAuthClient(&http.Client{Transport: rt})

		rt.RoundTripReturns(&http.Response{StatusCode: http.StatusUnauthorized}, nil)

		Expect(gh.ValidateToken(context.Background(), "sometoken")).To(HaveOccurred())
	})
	It("does not return an error when a token is valid", func() {
		rt := &fakehttp.FakeRoundTripper{}
		gh := NewGithubAuthClient(&http.Client{Transport: rt})
		rt.RoundTripReturns(&http.Response{StatusCode: http.StatusOK}, nil)

		Expect(gh.ValidateToken(context.Background(), "sometoken")).NotTo(HaveOccurred())
	})
})

type mockAuthRoundTripper struct {
	fn       func(r *http.Request) (*http.Response, error)
	calls    int
	callChan chan time.Time
}

func (rt *mockAuthRoundTripper) MockRoundTrip(fn func(r *http.Request) (*http.Response, error)) {
	rt.fn = fn
}

func (rt *mockAuthRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt.fn(r)
}

func newMockRoundTripper(pollCount int, token string, now func() time.Time) *mockAuthRoundTripper {
	rt := &mockAuthRoundTripper{calls: 0, callChan: make(chan time.Time, pollCount+1)}

	rt.MockRoundTrip(func(r *http.Request) (*http.Response, error) {
		b := bytes.NewBuffer(nil)

		var data githubAuthResponse

		switch {
		case rt.calls > pollCount:
			panic("mock API called after successful request")
		case rt.calls == pollCount:
			data = githubAuthResponse{AccessToken: token}
			rt.callChan <- now()
			close(rt.callChan)
		default:
			data = githubAuthResponse{Error: "slow_down"}
			rt.callChan <- now()
		}

		if err := json.NewEncoder(b).Encode(data); err != nil {
			return nil, err
		}

		res := &http.Response{
			Body: io.NopCloser(b),
		}

		res.StatusCode = http.StatusOK

		rt.calls = rt.calls + 1
		return res, nil
	})

	return rt
}
