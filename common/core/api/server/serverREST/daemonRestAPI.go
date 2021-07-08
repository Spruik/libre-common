package serverREST

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/utilities"
	"github.com/Spruik/libre-common/common/version"
	"github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
	"github.com/gorilla/mux"
	"log"
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
	s.router.HandleFunc("/", s.rootLink)

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

	for cmd, _ := range s.monitoredDaemon.GetCommands() {
		ep = fmt.Sprintf("/%s/control/{cmd}", s.monitoredDaemon.GetName())
		if cmd.GetInputParamNames() != nil {
			for _, p := range cmd.GetInputParamNames() {
				ep += fmt.Sprintf("/{%s}", p)
			}
		}
		s.router.HandleFunc(ep, s.controlCmdLink)
		s.endpoints = append(s.endpoints, ep)

	}
	ep = fmt.Sprintf("/%s/control/{cmd}", s.monitoredDaemon.GetName())
	s.router.HandleFunc(ep, s.controlCmdLink)
	s.endpoints = append(s.endpoints, ep)

	//set up the Kubernetes entry points - note: not adding these to the entrypoint list because they would not be called by a user
	s.router.HandleFunc("/home", s.homeLink)
	s.router.HandleFunc("/readyz", s.readyzLink)
	s.router.HandleFunc("/healthz", s.healthzLink)

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

func (s *DaemonRESTServer) rootLink(w http.ResponseWriter, r *http.Request) {
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
					topResp := map[string]interface{}{s.monitoredDaemon.GetName(): resp}
					var respBytes []byte
					respBytes, err = json.MarshalIndent(topResp, "", "   ")
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

///////////////////////////////////////////////////////////////////////////////////
//kubernetes REST API requirements
func (s *DaemonRESTServer) homeLink(w http.ResponseWriter, r *http.Request) {
	info := struct {
		BuildTime string `json:"buildTime"`
		Commit    string `json:"commit"`
		Release   string `json:"release"`
	}{
		version.BuildTime, version.Commit, version.Release,
	}

	body, err := json.Marshal(info)
	if err != nil {
		log.Printf("Could not encode info data: %v", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (s *DaemonRESTServer) healthzLink(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *DaemonRESTServer) readyzLink(w http.ResponseWriter, r *http.Request) {
	resp, err := s.monitoredDaemon.SubmitCommand(utilities.DaemonGetStateCommand, nil)
	if err == nil {
		anyInitializing := false
		for _, val := range resp {
			switch val.(type) {
			case string:
				status := fmt.Sprintf("%s", val)
				if status == utilities.DaemonInitialState.GetStateName() {
					anyInitializing = true
					break
				}
			}
		}
		if anyInitializing {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}
