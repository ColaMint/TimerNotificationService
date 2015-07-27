package main

const (
	ChannelBufferSize = 1000
)

type ChanGroup struct {
	chans       []chan interface{}
	chanCnt     int
	chanCounter int
}

func NewChanGroup(chanCnt int) *ChanGroup {
	chans := make([]chan interface{}, chanCnt)
	for i := 0; i < chanCnt; i++ {
		chans[i] = make(chan interface{}, ChannelBufferSize)
	}
	return &ChanGroup{chans: chans, chanCnt: chanCnt, chanCounter: -1}
}

func (this *ChanGroup) NextChan() chan interface{} {
	if this.chanCnt <= 0 {
		return nil
	}
	nextIndex := (this.chanCounter + 1) % this.chanCnt
	nextChan := this.chans[nextIndex]
	this.chanCounter = nextIndex
	return nextChan
}
