package shared

type SetArgs struct {
	Key, Value string
}

type SetReply struct {
	Exists bool
}

type KeyArg struct {
	Key string
}

type GetReply struct {
	Value string
	Ok    bool
}

type ListReply struct {
	Values []string
}

type EmptyArgs struct{}
