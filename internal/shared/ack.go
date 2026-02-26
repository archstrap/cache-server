package shared

type AckState struct {
	communication chan int
	replication   int
}

var ackState *AckState

func InitAckState(noOfReplicas int) {
	ackState = &AckState{
		communication: make(chan int, noOfReplicas),
		replication:   noOfReplicas,
	}
}

func GeAckState() *AckState {
	return ackState
}

func (a *AckState) GetCommunication() chan int {
	return a.communication
}
