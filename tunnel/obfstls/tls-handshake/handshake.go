package tlshs

import (
	"crypto/rand"
	"crypto/tls"
	"github.com/ninepeach/n4fd/log"
	"net"
	"time"
)

func ReadClientHello(c net.Conn) (sessionId []byte, err error) {

	record := &Record{}
	if _, err := record.ReadFrom(c); err != nil {
		log.Debug(err)
		return nil, err
	}

	if record.Type != RecordTypeHandshake {
		log.Info("readClientHello type: ", record.Type)
		return nil, ErrBadRecordType
	}

	if record.Version == tls.VersionTLS12 {
		log.Info("readClientHello type: ", record.Type, " version(tls12):", record.Version)
	} else if record.Version == tls.VersionTLS13 {
		log.Info("readClientHello type: ", record.Type, " version(tls13):", record.Version)
	} else {
		log.Info("readClientHello type: ", record.Type, " version(tls11 or 10):", record.Version)
	}

	clientMsg := &ClientHelloMsg{}
	if err := clientMsg.Decode(record.Opaque); err != nil {
		log.Debug(err)
		return nil, err
	}

	//log.Info("clientMsg Version:", clientMsg.Version)
	//log.Info("clientMsg Random:", clientMsg.Random)
	//log.Info("clientMsg SessionId:", string(clientMsg.SessionID))

	//for _, c := range clientMsg.CipherSuites {
	//	log.Info("clientMsg CipherSuites:", tls.CipherSuiteName(c), c)
	//}

	if len(clientMsg.SessionID) == 0 {
		return nil, ErrBadSessionIdType
	}

	/*
		for _, ext := range clientMsg.Extensions {
			log.Info("clientMsg ExtSession:", ext.Type())
			if ext.Type() == ExtSessionTicket {
				log.Info("clientMsg ExtSession is ExtSessionTicket", ext.Type())
				b, err := ext.Encode()
				if err != nil {
					log.Debug(err)
					return clientMsg.SessionID, err
				}
				c.Write(b)
				break
			}
		}
	*/
	return clientMsg.SessionID, nil
}

func ReadAppData(c net.Conn) (err error) {

	record := &Record{}
	if _, err := record.ReadFrom(c); err != nil {
		log.Debug(err)
		return err
	}

	if record.Type != RecordTypeAppData {
		log.Info("readAppData type: ", record.Type)
		return ErrBadRecordType
	}

	log.Info("readAppData type: ", record.Type, " version:", record.Version)

	//record.Opaque
	return nil
}

func SendAlertProtocol(c net.Conn) error {

	record := &Record{
		Type:    RecordTypeAlert,
		Version: tls.VersionTLS10,
		Opaque:  []byte{2, 70},
	}
	record.WriteTo(c)
	return nil
}

func SendServerHello(c net.Conn, SessionID []byte) error {

	log.Info("sendServerHello go ")

	serverMsg := &ServerHelloMsg{
		Version:           tls.VersionTLS12,
		SessionID:         SessionID,
		CipherSuite:       0xcca8,
		CompressionMethod: 0x00,
		Extensions: []Extension{
			&RenegotiationInfoExtension{},
			&ExtendedMasterSecretExtension{},
			&ECPointFormatsExtension{
				Formats: []uint8{0x00},
			},
		},
	}

	serverMsg.Random.Time = uint32(time.Now().Unix())
	rand.Read(serverMsg.Random.Opaque[:])

	b, err := serverMsg.Encode()
	if err != nil {
		log.Info("encode failed ", err)

		return err
	}

	record := &Record{
		Type:    RecordTypeHandshake,
		Version: tls.VersionTLS10,
		Opaque:  b,
	}
	_, err = record.WriteTo(c)

	return err
}
func serverHandshake(c net.Conn) error {
	sessionId, err := ReadClientHello(c)
	if err == nil {
		log.Info("sendServerHello")
		SendServerHello(c, sessionId)
	}

	err = ReadAppData(c)
	log.Info("readAppData ", err)
	return nil
	/*
		serverMsg := &ServerHelloMsg{
			Version:           tls.VersionTLS12,
			SessionID:         clientMsg.SessionID,
			CipherSuite:       0xcca8,
			CompressionMethod: 0x00,
			Extensions: []Extension{
				&RenegotiationInfoExtension{},
				&ExtendedMasterSecretExtension{},
				&ECPointFormatsExtension{
					Formats: []uint8{0x00},
				},
			},
		}

		log.Debug("write serverMsg")

		serverMsg.Random.Time = uint32(time.Now().Unix())
		rand.Read(serverMsg.Random.Opaque[:])
		b, err := serverMsg.Encode()
		if err != nil {
			return err
		}

		record = &Record{
			Type:    Handshake,
			Version: tls.VersionTLS10,
			Opaque:  b,
		}

		if _, err := record.WriteTo(bufio.NewWriter(c)); err != nil {
			return err
		}

		record = &Record{
			Type:    ChangeCipherSpec,
			Version: tls.VersionTLS12,
			Opaque:  []byte{0x01},
		}
		if _, err := record.WriteTo(bufio.NewWriter(c)); err != nil {
			return err
		}
	*/
	return nil
}
