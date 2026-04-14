package tcp

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/liftel/epic-fain/internal/domain/model"
	"github.com/liftel/epic-fain/internal/domain/port"
)

// FrameProtocol defines the TCP wire format for CAN frames:
//
//	[2 bytes] installation ID length (big-endian uint16)
//	[N bytes] installation ID (UTF-8)
//	[2 bytes] CAN message ID (big-endian uint16)
//	[1 byte]  DLC
//	[DLC bytes] CAN data payload
const maxInstallationIDLen = 256

// Server receives CAN frames over TCP and feeds them to the telemetry service.
type Server struct {
	listenAddr   string
	telemetrySvc port.TelemetryService
	listener     net.Listener
}

func NewServer(addr string, telemetrySvc port.TelemetryService) *Server {
	return &Server{
		listenAddr:   addr,
		telemetrySvc: telemetrySvc,
	}
}

// Start begins accepting TCP connections. Blocks until context is cancelled.
func (s *Server) Start(ctx context.Context) error {
	var err error
	s.listener, err = net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", s.listenAddr, err)
	}
	log.Printf("[TCP] Listening on %s", s.listenAddr)

	go func() {
		<-ctx.Done()
		s.listener.Close()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("[TCP] Accept error: %v", err)
				continue
			}
		}
		go s.handleConnection(ctx, conn)
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()
	log.Printf("[TCP] New connection from %s", remoteAddr)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		// Read installation ID length
		var idLen uint16
		if err := binary.Read(conn, binary.BigEndian, &idLen); err != nil {
			if err != io.EOF {
				log.Printf("[TCP] Error reading ID length from %s: %v", remoteAddr, err)
			}
			return
		}
		if idLen > maxInstallationIDLen {
			log.Printf("[TCP] Installation ID too long (%d) from %s", idLen, remoteAddr)
			return
		}

		// Read installation ID
		idBuf := make([]byte, idLen)
		if _, err := io.ReadFull(conn, idBuf); err != nil {
			log.Printf("[TCP] Error reading installation ID from %s: %v", remoteAddr, err)
			return
		}
		installationID := string(idBuf)

		// Read CAN message ID
		var msgID uint16
		if err := binary.Read(conn, binary.BigEndian, &msgID); err != nil {
			log.Printf("[TCP] Error reading message ID from %s: %v", remoteAddr, err)
			return
		}

		// Read DLC
		var dlc uint8
		if err := binary.Read(conn, binary.BigEndian, &dlc); err != nil {
			log.Printf("[TCP] Error reading DLC from %s: %v", remoteAddr, err)
			return
		}
		if dlc > 8 {
			log.Printf("[TCP] Invalid DLC %d from %s", dlc, remoteAddr)
			return
		}

		// Read CAN data
		data := make([]byte, dlc)
		if _, err := io.ReadFull(conn, data); err != nil {
			log.Printf("[TCP] Error reading CAN data from %s: %v", remoteAddr, err)
			return
		}

		frame := model.CANFrame{
			MessageID:  model.MessageID(msgID),
			DLC:        dlc,
			Data:       data,
			ReceivedAt: time.Now(),
		}

		if err := s.telemetrySvc.IngestFrame(ctx, installationID, frame); err != nil {
			log.Printf("[TCP] Error ingesting frame 0x%04X for %s: %v", msgID, installationID, err)
		}
	}
}
