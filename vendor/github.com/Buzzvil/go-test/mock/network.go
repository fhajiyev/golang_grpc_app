package mock

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

//ClientPatcher is a helper structure which will contains the client will be patched, the original transport and mock transport.
type ClientPatcher struct {
	OriginalTransport http.RoundTripper
	MockTransport     http.RoundTripper
	httpClient        *http.Client
}

//PatchClient will replace the original transport with mock transport.
//It will return ClientPatcher for remove the mock transport after using it.
func PatchClient(httpClient *http.Client, targetServers ...*TargetServer) *ClientPatcher {

	t := &transportCarrier{
		OriginalTransport: httpClient.Transport,
		TargetServers:     targetServers,
	}

	if t.OriginalTransport == nil {
		t.OriginalTransport = &http.Transport{}
	}

	patcher := &ClientPatcher{
		OriginalTransport: httpClient.Transport,
		MockTransport:     t,
		httpClient:        httpClient,
	}

	httpClient.Transport = patcher.MockTransport

	return patcher
}

//NetTargetServer will returns TargetServer with the host.
func NewTargetServer(host string) *TargetServer {
	if host == "" {
		panic("Host should be defined.")
	}

	return &TargetServer{
		Host: host,
	}
}

//AddResponseHandler will add ResponseHandler to the TargetServer.
func (s *TargetServer) AddResponseHandler(r *ResponseHandler) *TargetServer {
	if r.WriteToBody == nil {
		panic("WriteToBody of resHandler should be defined.")
	}

	if r.StatusCode == 0 {
		r.StatusCode = 200
	}

	if r.Method == "" {
		panic("Method of resHandler should be defined.")
	}

	s.ResponseHandlers = append(s.ResponseHandlers, r)
	return s
}

//RemovePatch will replace the mock transport with original.
func (httpReq *ClientPatcher) RemovePatch() {
	httpReq.httpClient.Transport = httpReq.OriginalTransport
}

type TargetServer struct {
	Host             string
	ResponseHandlers []*ResponseHandler
}

//ResponseHandler contains a mock response and a request should be handled.
type ResponseHandler struct {
	WriteToBody func() []byte
	StatusCode  int
	Path        string
	Method      string
}

type transportCarrier struct {
	TargetServers     []*TargetServer
	OriginalTransport http.RoundTripper
}

func (m *transportCarrier) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, targetServer := range m.TargetServers {
		if targetServer.Host == req.Host {
			for _, resHandler := range targetServer.ResponseHandlers {
				if resHandler.Path == req.URL.Path && resHandler.Method == req.Method {
					// mock response object
					r := http.Response{
						StatusCode: resHandler.StatusCode,
						Proto:      "HTTP/1.0",
						ProtoMajor: 1,
						ProtoMinor: 0,
					}

					buf := bytes.NewBuffer(resHandler.WriteToBody())
					r.Body = ioutil.NopCloser(buf)

					return &r, nil
				}
			}
		}
	}
	return m.OriginalTransport.RoundTrip(req)
}
