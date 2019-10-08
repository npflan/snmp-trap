package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"go.uber.org/zap"

	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/types"
	snmp "github.com/soniah/gosnmp"
)

type SnmpService struct {
	logger *zap.Logger
}

func (s *SnmpService) getSubtree(oidString string) ([]gosmi.SmiNode, error) {
	oid, err := types.OidFromString(oidString)

	if err != nil {
		s.logger.Error(
			"Failed to get OID from octet variable name",
			zap.Error(err),
		)
		return nil, err
	}

	node, err := gosmi.GetNodeByOID(oid)
	if err != nil {
		s.logger.Error(
			"Failed to get smnp node by OID",
			zap.Error(err),
		)
		return nil, err
	}
	return node.GetSubtree(), nil
}

func (s *SnmpService) myTrapHandler(packet *snmp.SnmpPacket, addr *net.UDPAddr) {
	var logFieldsMap = make(map[string]*string)
	var logFields []zap.Field
	for _, v := range packet.Variables {
		switch v.Type {
		case snmp.ObjectIdentifier:
			// Get OID Value from type OID.
			value := v.Value.(string)
			subtree, err := s.getSubtree(value)
			if err != nil {
				continue
			}
			// Collect info to build logline from.
			oidName := subtree[0].Node.Name
			logFieldsMap["TrapName"] = &oidName
			oidDescription := subtree[0].Node.Description
			logFieldsMap["TrapDescription"] = &oidDescription

		case snmp.OctetString:
			value := v.Value.([]byte)
			subtree, err := s.getSubtree(v.Name)
			if err != nil {
				continue
			}
			// Collect info to build logline.
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
	// Add source addr to log lines
	sourceAddr := addr.IP.String()
	logFieldsMap["TrapSourceAddr"] = &sourceAddr
	for k, v := range logFieldsMap {
		logFields = append(logFields, zap.String(k, *v))
	}
	s.logger.Info(
		"Received trap",
		logFields...,
	)
}

func main() {
	gosmi.Init()
	gosmi.AppendPath("./mibs")

	files, _ := filepath.Glob("./mibs/*")

	for _, filePath := range files {
		_, fileName := path.Split(filePath)
		_, err := gosmi.LoadModule(fileName)
		if err != nil {
			fmt.Printf("failed to load %s", fileName)
			continue
		}
	}
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	zapLogger, err := config.Build()
	if err != nil {
		panic(err)
	}
	svc := SnmpService{
		logger: zapLogger,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	tl := snmp.NewTrapListener()
	tl.OnNewTrap = svc.myTrapHandler
	tl.Params = snmp.Default
	go func() {
		err = tl.Listen("0.0.0.0:162")
		if err != nil {
			panic(err)
		}
	}()
	<-done
	_ = zapLogger.Sync()
}
