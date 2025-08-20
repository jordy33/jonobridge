package models

import "encoding/xml"

type GetReturnMessagesResult struct {
	XMLName      xml.Name     `xml:"GetReturnMessagesResult"`
	ErrorId      uint64       `xml:"ErrorID"`
	More         bool         `xml:"More"`
	NextStartUTC NextStartUTC `xml:"NextStartUTC"`
	NextStartID  uint64       `xml:"NextStartID"`
	Messages     Messages     `xml:"Messages"`
}

type NextStartUTC struct {
	XMLName xml.Name `xml:"NextStartUTC"`
}

type Messages struct {
	XMLName          xml.Name           `xml:"Messages"`
	ReturnedMessages []ReturnedMessages `xml:"ReturnMessage"`
}

type ReturnedMessages struct {
	XMLName        xml.Name `json:"xmlname" xml:"ReturnMessage"`
	ID             uint64   `json:"id" xml:"ID"`
	MessageUTC     string   `json:"messageUTC" xml:"MessageUTC"`
	ReceiveUTC     string   `json:"receiveUTC" xml:"ReceiveUTC"`
	SIN            int64    `json:"sin" xml:"ReturnMessage"`
	MobileID       string   `json:"mobileid" xml:"MobileID"`
	Payload        Payload  `json:"payload" xml:"Payload,omitempty"`
	RegionName     string   `json:"regionmame" xml:"RegionName"`
	OtaMessageSize string   `json:"otamessagesize" xml:"OTAMessageSize"`
}

type Payload struct {
	XMLName xml.Name `json:"xmlname" xml:"Payload"`
	Name    string   `json:"name" xml:"Name,attr"`
	Sin     string   `json:"sin" xml:"SIN,attr"`
	Min     string   `json:"min" xml:"MIN,attr"`
	Fields  Fields   `json:"fields" xml:"Fields"`
}

type Fields struct {
	XMLName xml.Name `json:"xmlname" xml:"Fields"`
	Fields  []Field  `json:"fields" xml:"Field"`
}

type Field struct {
	XMLName xml.Name `json:"xmlname" xml:"Field"`
	Name    string   `json:"name" xml:"Name,attr"`
	Value   string   `json:"value" xml:"Value,attr"`
}

type PayloadBridge struct {
	ID             uint64 `json:"id"`
	MessageUTC     string `json:"messageUTC"`
	ReceiveUTC     string `json:"receiveUTC"`
	SIN            int64  `json:"sin"`
	MobileID       string `json:"mobileid"`
	RegionName     string `json:"regionname"`
	OtaMessageSize string `json:"otamessagesize"`
	Type           string `json:"type"`
	Min            string `json:"min"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	Speed          string `json:"speed"`
	Heading        string `json:"heading"`
	EventTime      string `json:"eventtime"`
	GpsFixAge      string `json:"gpsfixage"`
}
