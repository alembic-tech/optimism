package main

import "net/http"

type aggregator struct {}

func newAggregator() *aggregator {
  return &aggregator{}
}

func (m *aggregator) handleGet(w http.ResponseWriter, req *http.Request) {
}

func (m *aggregator) handlePost(w http.ResponseWriter, req *http.Request) {
}
