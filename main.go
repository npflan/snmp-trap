package main

import (
	"fmt"
	"log"
	"net"

	"go.uber.org/zap"

	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/types"
	snmp "github.com/soniah/gosnmp"
)

type SnmpService struct {
	logger *zap.Logger
}

func main() {
	// _, err := parser.ParseFile("CISCO-SMI.my")
	gosmi.Init()
	gosmi.AppendPath("./mibs")
	_, err := gosmi.LoadModule("CISCO-VTP-MIB.my")
	if err != nil {
		panic(err)
	}
	svc := SnmpService{
		logger: zap.NewExample(),
	}

	tl := snmp.NewTrapListener()
	tl.OnNewTrap = svc.myTrapHandler
	tl.Params = snmp.Default
	err = tl.Listen("0.0.0.0:162")
	if err != nil {
		log.Panicf("error in listen: %s", err)
	}
}

func (s *SnmpService) myTrapHandler(packet *snmp.SnmpPacket, addr *net.UDPAddr) {
	var logFieldsMap = make(map[string]*string)
	var logFields []zap.Field
	sourceAddr := addr.String()
	logFieldsMap["TrapSourceAddr"] = &sourceAddr
	for _, v := range packet.Variables {
		switch v.Type {
		case snmp.ObjectIdentifier:
			// Get OID Value from type OID.
			value := v.Value.(string)
			oid, err := types.OidFromString(value)
			if err != nil {
				fmt.Println("shit")
			}
			node, err := gosmi.GetNodeByOID(oid)
			if err != nil {
				fmt.Println("shit 1")
			}
			subtree := node.GetSubtree()
			// Collect info to build logline from.
			oidName := subtree[0].Node.Name
			logFieldsMap["TrapName"] = &oidName
			oidDescription := subtree[0].Node.Description
			logFieldsMap["TrapDescription"] = &oidDescription

		case snmp.OctetString:
			value := v.Value.([]byte)
			oid, err := types.OidFromString(v.Name)
			if err != nil {
				s.logger.Error(
					"Failed to get OID from octet variable name",
					zap.Error(err),
				)
			}
			node, err := gosmi.GetNodeByOID(oid)
			if err != nil {
				s.logger.Error(
					"Failed to get OID from octet variable name",
					zap.Error(err),
				)
			}
			subtree := node.GetSubtree()
			octetOidName := subtree[0].Node.Name
			octetDescription := subtree[0].Node.Description
			octetValue := string(value)
			logFieldsMap["TrapEventName"] = &octetOidName
			logFieldsMap["TrapEventDescription"] = &octetDescription
			logFieldsMap["TrapEventValue"] = &octetValue
		case snmp.TimeTicks:
			continue
		default:
			continue
		}
	}
	for k, v := range logFieldsMap {
		logFields = append(logFields, zap.String(k, *v))
	}
	s.logger.Info(
		"Received trap",
		logFields...,
	)
}
