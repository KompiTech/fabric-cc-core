package main

import "github.com/KompiTech/rmap"

// this is potential index json data to be saved
type IndexFile struct {
	data     rmap.Rmap
	fileName string
}

type MultiIndex struct {
	fieldName string
}
