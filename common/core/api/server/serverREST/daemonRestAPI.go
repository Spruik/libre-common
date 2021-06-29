package serverREST

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
	"strings"
)

type DaemonRESTServer struct {
	libreConfig.ConfigurationEnabler
	libreLogger.LoggingEnabler

	monitoredDaemon ports.DaemonIF
	router          *mux.Router

	endpoints []string

	port       string
	httpServer *http.Server
}

func NewDaemonRESTServer(daemon ports.DaemonIF) *DaemonRESTServer {
	s := DaemonRESTServer{
		monitoredDaemon: daemon,
		endpoints:       make([]string, 0, 0),
	}
	s.SetLoggerConfigHook("RESTAPI")
	s.SetConfigCategory("RESTAPI")
	s.router = mux.NewRouter().StrictSlash(true)
	s.router.HandleFunc("/", s.homeLink)

	eps := libreLogger.GetRESTAPIEntryPoints()
	for ep, f := range eps {
		s.router.HandleFunc(ep, f)
		s.endpoints = append(s.endpoints, ep)
	}

	sort.Strings(s.endpoints)
	var portStr string
	var err error
	portStr, err = s.GetConfigItemWithDefault("PORT", "8080")
	if err == nil {
		s.port = portStr
	} else {
		s.LogErrorf("config error looking for api port - defaulting to 8080 [%s]", err)
	}

	//set up the daemon control functions entry points
	ep := fmt.Sprintf("/%s/control", s.monitoredDaemon.GetName())
	s.router.HandleFunc(ep, s.controlLink)
	s.endpoints = append(s.endpoints, ep)

	ep = fmt.Sprintf("/%s/control/{cmd}", s.monitoredDaemon.GetName())
	s.router.HandleFunc(ep, s.controlCmdLink)
	s.endpoints = append(s.endpoints, ep)

	return &s
}

func (s *DaemonRESTServer) Start() error {
	s.httpServer = &http.Server{
		Addr:    ":" + s.port,
		Handler: s.router,
	}
	return s.httpServer.ListenAndServe()
}

func (s *DaemonRESTServer) Shutdown() error {
	return s.httpServer.Shutdown(context.Background())
}

func (s *DaemonRESTServer) homeLink(w http.ResponseWriter, r *http.Request) {
	s.LogDebugf("handling request to homelink from user agent %s", r.UserAgent())
	var resp = struct {
		DaemonName string
		Endpoints  []string
	}{
		DaemonName: s.monitoredDaemon.GetName(),
		Endpoints:  make([]string, 0, 0),
	}
	for _, ep := range s.endpoints {
		resp.Endpoints = append(resp.Endpoints, ep)
	}
	respBytes, err := json.Marshal(&resp)
	if err == nil {
		_, _ = fmt.Fprintln(w, string(respBytes))
	} else {
		_, _ = fmt.Fprintln(w, fmt.Sprintf("%s", err))
	}
}

func (s *DaemonRESTServer) controlLink(w http.ResponseWriter, r *http.Request) {
	_ = r
	respStr := "serverREST API for: " + s.monitoredDaemon.GetName() + " - control"
	respStr += "   Available endpoints are: \n"
	for cmd, _ := range s.monitoredDaemon.GetCommands() {
		respStr += fmt.Sprintf("      /%s/control/%s\n", s.monitoredDaemon.GetName(), cmd.GetCommandName())
	}
	_, _ = fmt.Fprintln(w, respStr)
}

func (s *DaemonRESTServer) controlCmdLink(w http.ResponseWriter, r *http.Request) {
	cmdName := mux.Vars(r)["cmd"]
	var targetCommand ports.DaemonCommandIF = nil
	for cmd, _ := range s.monitoredDaemon.GetCommands() {
		if strings.ToUpper(cmd.GetCommandName()) == strings.ToUpper(cmdName) {
			targetCommand = cmd
			break
		}
	}
	if targetCommand != nil {
		err := r.ParseForm()
		if err == nil {
			params := make(map[string]interface{})
			for i, j := range r.Form {
				if len(j) == 1 {
					params[i] = j[0]
				} else {
					params[i] = j
				}
			}
			resp, err := s.monitoredDaemon.SubmitCommand(targetCommand, params)
			if err == nil {
				if resp != nil {
					var respBytes []byte
					respBytes, err = json.Marshal(resp)
					if err == nil {
						_, _ = fmt.Fprintln(w, string(respBytes))
					}
				} else {
					_, _ = fmt.Fprintln(w, "Command completed successfully with no return data")
				}
			} else {
				_, _ = fmt.Fprintln(w, fmt.Sprintf("Error executing command:%+v", err))
			}
		}
	}

}
