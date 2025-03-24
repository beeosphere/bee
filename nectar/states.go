package nectar

// import "sync"

// type ProcessorState int

// const (
// 	Stopped ProcessorState = iota
// 	Started
// 	Executing
// 	Failed
// )

// var (
// 	stateLock sync.Mutex
// 	currentState ProcessorState
// )

// func LockState() {
// 	stateLock.Lock()
// }

// func UnlockState() {
// 	stateLock.Unlock()
// }

// func GetState() ProcessorState {
// 	stateLock.Lock()
// 	defer stateLock.Unlock()
// 	return currentState
// }

// func SetState(state ProcessorState) {
// 	stateLock.Lock()
// 	defer stateLock.Unlock()
// 	currentState = state
// }
