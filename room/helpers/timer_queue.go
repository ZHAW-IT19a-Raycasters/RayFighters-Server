package helpers

import "time"

var getTime = time.Now

type TimerData struct {
	id   uint8
	from time.Time
}

func (td *TimerData) GetId() uint8 {
	return td.id
}

func (td *TimerData) GetTime() time.Time {
	return td.from
}

func newTimeData(id uint8) TimerData {
	return TimerData{
		id:   id,
		from: getTime(),
	}
}

type ClientTimer struct {
	id     uint8
	sender chan<- TimerData
}

func (ct ClientTimer) RegisterTimeout() {
	ct.sender <- newTimeData(ct.id)
}

func newClientTimer(id uint8, sender chan<- TimerData) ClientTimer {
	return ClientTimer{
		id:     id,
		sender: sender,
	}
}

type HubTimer struct {
	queue chan TimerData
}

func NewHubTimer(size int) HubTimer {
	return HubTimer{
		queue: make(chan TimerData, size),
	}
}

func (ht HubTimer) GetQueue() <-chan TimerData {
	return ht.queue
}

func (ht HubTimer) NewClientTimer(id uint8) ClientTimer {
	return newClientTimer(id, ht.queue)
}
