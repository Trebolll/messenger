package model

type TypeChat string

const (
	TypePrivate TypeChat = "private"
	TypePublic  TypeChat = "public"
	TypeGroup   TypeChat = "group"
	TypeChannel TypeChat = "channel"
)
