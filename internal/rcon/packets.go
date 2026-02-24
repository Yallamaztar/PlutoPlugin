package rcon

import (
	"bytes"
	"fmt"
	"time"
)

func (r *RCON) buildPacket(command string, needsPrefix bool) []byte {
	var payload string
	if needsPrefix {
		payload = fmt.Sprintf("rcon %s %s", r.password, command)
	} else {
		payload = command
	}

	packet := append([]byte{0xFF, 0xFF, 0xFF, 0xFF}, []byte(payload)...)
	packet = append(packet, '\n')
	return packet
}

func (r *RCON) sendPacket(packet []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error
	for attempt := 1; attempt <= defaultRetryCount; attempt++ {
		_, err := r.conn.Write(packet)
		if err == nil {
			return nil
		}
		lastErr = fmt.Errorf("send attempt %d/%d failed: %w", attempt, defaultRetryCount, err)
	}
	return lastErr
}

func (r *RCON) readResponse() ([]string, error) {
	var buf bytes.Buffer
	deadline := time.Now().Add(defaultReadTimeout)

	for {
		if err := r.conn.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}

		tmp := make([]byte, readBufferSize)
		n, err := r.conn.Read(tmp)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		buf.Write(tmp[:n])
		if n < readBufferSize {
			break
		} else {
			deadline = time.Now().Add(defaultReadExtension)
		}
	}

	raw := normalizeRCON(buf.String())
	lines := splitNonEmptyLines(raw)
	if len(lines) > 0 {
		return lines, nil
	}

	return lines, nil
}
