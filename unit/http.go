package unit

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/giantswarm/nomi/log"
)

type UnitObserver struct {
	unitEngine *UnitEngine
}

func NewUnitObserver(engine *UnitEngine) *UnitObserver {
	return &UnitObserver{
		unitEngine: engine,
	}
}

func (s *UnitObserver) StartHTTPService(addr string) {
	r := mux.NewRouter()
	r.HandleFunc("/hello/{unitID}", withIDParam(s.HelloHandler)).Methods("GET")
	r.HandleFunc("/alive/{unitID}", withIDParam(s.AliveHandler)).Methods("GET")
	r.HandleFunc("/bye/{unitID}", withIDParam(s.ByeHandler)).Methods("GET")

	r.HandleFunc("/stats/{statsID}", s.StatsHandler).Methods("POST")

	http.Handle("/", r)
	if Verbose {
		log.Logger().Infof("listening on %s\n", addr)
	}

	go http.ListenAndServe(addr, nil)
	if Verbose {
		log.Logger().Infof("listening on %s\n", addr)
	}
}

func withIDParam(handler func(unitID string, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		unitID := mux.Vars(r)["unitID"]
		if unitID == "" {
			log.Logger().Error("empty unitID")
			w.WriteHeader(400)
		} else {
			handler(unitID, w, r)
		}
	}
}

func (s *UnitObserver) HelloHandler(unitID string, w http.ResponseWriter, r *http.Request) {
	delay := s.unitEngine.MarkUnitRunning(unitID)
	if Verbose {
		log.Logger().Infof("marked unit as running: %s [%d] %f", unitID, len(s.unitEngine.runningUnits), delay.Seconds())
	}
	w.Write([]byte("ok.\n"))
}

func (s *UnitObserver) AliveHandler(unitID string, w http.ResponseWriter, r *http.Request) {
	if _, isStopped := s.unitEngine.stoppingUnits[unitID]; isStopped {
		w.WriteHeader(500)
	} else {
		w.Write([]byte("ok.\n"))
	}
}
func (s *UnitObserver) ByeHandler(unitID string, w http.ResponseWriter, r *http.Request) {
	if Verbose {
		log.Logger().Infof("marking unit as stopped: %s [%d]", unitID, len(s.unitEngine.stoppedUnits)+len(s.unitEngine.runningUnits)+len(s.unitEngine.startingUnits))
	}
	s.unitEngine.MarkUnitStopped(unitID)
	w.Write([]byte("ok.\n"))
}

func (s *UnitObserver) StatsHandler(w http.ResponseWriter, r *http.Request) {
	statsID := mux.Vars(r)["statsID"]
	b := bytes.NewBufferString("")
	b.ReadFrom(r.Body)

	var hostname string
	var cpuusage float64
	var rss int

	n, err := fmt.Sscanf(b.String(), "%s %f %d", &hostname, &cpuusage, &rss)
	if err != nil || n != 3 {
		log.Logger().Warningf("don't know how to parse statsline: " + b.String())
		return
	}

	s.unitEngine.DumpProcessStats(statsID, hostname, cpuusage, rss)
}
