package main

import (
	"encoding/json"
	"net/http"

	"github.com/cockroachdb/pebble"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type server struct {
	db   *pebble.DB
	port string
}

func newServer(database string, port string) (*server, error) {
	s := server{db: nil, port: port}

	var err error
	s.db, err = pebble.Open(database, &pebble.Options{})

	return &s, err
}

func (s *server) addDocument(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	dec := json.NewDecoder(r.Body)
	var document map[string]any
	err := dec.Decode(&document)

	if err != nil {
		jsonResponse(w, nil, err)
		return
	}

	id := uuid.New().String()

	bs, err := json.Marshal(document)
	if err != nil {
		jsonResponse(w, nil, err)
		return
	}

	err = s.db.Set([]byte(id), bs, pebble.Sync)
	if err != nil {
		jsonResponse(w, nil, err)
		return
	}

	jsonResponse(w, map[string]any{"id": id}, nil)

}

func (s *server) searchDocuments(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q, err := parseQuery(r.URL.Query().Get("q"))
	if err != nil {
		jsonResponse(w, nil, err)
		return
	}

	var documents []map[string]any

	iter, _ := s.db.NewIter(nil)
	defer iter.Close()
	for iter.First(); iter.Valid(); iter.Next() {
		var document map[string]any
		err = json.Unmarshal(iter.Value(), &document)
		if err != nil {
			jsonResponse(w, nil, err)
			return
		}

		if q.match(document) {
			documents = append(documents, map[string]any{
				"id":   string(iter.Key()),
				"body": document,
			})
		}
	}

	jsonResponse(w, map[string]any{"documents": documents, "count": len(documents)}, nil)

}

func (s *server) getDocumentById(id []byte) (map[string]any, error) {
	valBytes, closer, err := s.db.Get(id)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	var document map[string]any
	err = json.Unmarshal(valBytes, &document)

	return document, err
}

func (s *server) getDocument(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	document, err := s.getDocumentById([]byte(id))
	if err != nil {
		jsonResponse(w, nil, err)
		return
	}

	jsonResponse(w, map[string]any{"document": document}, nil)

}

func jsonResponse(w http.ResponseWriter, body map[string]any, err error) {
	data := map[string]any{
		"body":   body,
		"status": "ok",
	}

	if err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		data["status"] = "error"
		data["error"] = err.Error()
		w.WriteHeader(http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	err = enc.Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
