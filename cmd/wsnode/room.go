package main

import "sync"

type Topic struct {
	Owner  string
	Member sync.Map
}
