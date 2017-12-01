package client

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
)

// SplitManager provides information of the currently stored splits
type SplitManager struct {
	splitStorage storage.SplitStorageConsumer
}

// SplitView is a partial representation of a currently stored split
type SplitView struct {
	Name         string   `json:"name"`
	TrafficType  string   `json:"trafficType"`
	Killed       bool     `json:"killed"`
	Treatments   []string `json:"treatments"`
	ChangeNumber int64    `json:"changeNumber"`
}

func newSplitView(splitDto *dtos.SplitDTO) *SplitView {
	treatments := make([]string, 0)
	for _, condition := range splitDto.Conditions {
		for _, partition := range condition.Partitions {
			treatments = append(treatments, partition.Treatment)
		}
	}
	return &SplitView{
		ChangeNumber: splitDto.ChangeNumber,
		Killed:       splitDto.Killed,
		Name:         splitDto.Name,
		TrafficType:  splitDto.TrafficTypeName,
		Treatments:   treatments,
	}
}

// SplitNames returns a list with the name of all the currently stored splits
func (m *SplitManager) SplitNames() []string {
	return m.splitStorage.SplitNames()
}

// Splits returns a list of a partial view of every currently stored split
func (m *SplitManager) Splits() []SplitView {
	splitViews := make([]SplitView, 0)
	splits := m.splitStorage.GetAll()
	for _, split := range splits {
		splitViews = append(splitViews, *newSplitView(&split))
	}
	return splitViews
}

// Split returns a partial view of a particular split
func (m *SplitManager) Split(feature string) *SplitView {
	split := m.splitStorage.Get(feature)
	if split != nil {
		return newSplitView(split)
	}
	return nil
}