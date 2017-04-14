package main

// Structs for use with pogs and the grain.UiView_ViewInfo type.

type viewInfo struct {
	MatchRequests, MatchOffers []PowerboxDescriptor
}

type PowerboxDescriptor struct {
	Tags []Tag
}

type Tag struct {
	Id uint64
}
