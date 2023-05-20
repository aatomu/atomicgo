package discordbot

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/ogg"
)

type encodeSession struct {
	Status   EncodeStats
	filePath string
	opts     EncodeOpts

	flames         chan []byte
	isFlamesClosed bool
	vc             *discordgo.VoiceConnection
	process        *os.Process
	done           chan error
}

// FFmpeg Options
type EncodeOpts struct {
	FlameBuf    int     // default 100(20ms*100=2s)
	Volume      int     // volume 0-512(default:256)
	Offset      float64 // encode start time
	AudioFilter string  // ffmpeg Filters,https://ffmpeg.org/ffmpeg-filters.html#Audio-Filters
	Compression int     //higher is best quality & slow encoding(0-10)
}

type EncodeStats struct {
	Size    int
	Time    time.Duration
	Bitrate float32
	Speed   float32
}

func NewFileEncodeStream(v *discordgo.VoiceConnection, f string, opts EncodeOpts, done chan error) (s *encodeSession) {
	encodeSetting := EncodeOpts{
		FlameBuf:    100,
		Volume:      256,
		Offset:      opts.Offset,
		AudioFilter: opts.AudioFilter,
		Compression: opts.Compression,
	}
	if opts.FlameBuf != 0 {
		encodeSetting.FlameBuf = opts.FlameBuf
	}
	if opts.Volume != 0 {
		encodeSetting.Volume = opts.Volume
	}

	s = &encodeSession{
		filePath: f,
		opts:     encodeSetting,
		flames:   make(chan []byte, opts.FlameBuf),
		vc:       v,
		done:     done,
	}
	go s.run()
	go s.stream()

	return
}

func (s *encodeSession) run() {
	defer func() {
		if s.isFlamesClosed {
			return
		}
		close(s.flames)
		s.isFlamesClosed = true
	}()

	args := []string{
		// Default
		"-stats",
		"-stats_period", "0.1",
		"-i", s.filePath,
		"-vol", strconv.Itoa(s.opts.Volume),
		"-ss", fmt.Sprintf("%.2f", s.opts.Offset),
		"-compression_level", strconv.Itoa(s.opts.Compression),
		// Audio Encode Opts
		"-ab", "64k", // Bitrate
		"-ar", "48000", //Sampling Rate
		"-ac", "2", // Audio Channel
		"-acodec", "libopus", // Audio Codec
		// Other Encode Opts
		"-map", "0:a", // Output Stream Map
		"-f", "ogg", // Encode Format
		"-frame_duration", "20", // Frame Size
		"-packet_loss", "1", // Packet Lossing
		// Output(StdOut)
	}

	// check Audio Filter
	if s.opts.AudioFilter != "" {
		args = append(args, "-af", s.opts.AudioFilter)
	}

	args = append(args, "pipe:1")
	ffmpeg := exec.Command("ffmpeg", args...)

	s.process = ffmpeg.Process

	stderr, err := ffmpeg.StderrPipe()
	if err != nil {
		panic("get stdoutpipe error")
	}
	defer stderr.Close()
	go s.readStdErr(stderr)

	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		s.done <- fmt.Errorf("get stdoutpipe error")
		return
	}
	defer stdout.Close()

	ffmpeg.Start()
	s.readStdOut(stdout)

	ffmpeg.Wait()
}

func (s *encodeSession) readStdErr(stderr io.Reader) {
	scanner := bufio.NewReader(stderr)
	for {
		line, err := scanner.ReadString('\r')
		if err != nil {
			if err != io.EOF {
				log.Println("Error Reading stderr:", err)
			}
			break
		}
		if !strings.HasPrefix(line, "size=") || strings.Contains(line, "N/A") {
			continue // Not stats info
		}

		var timeHour, timeMin int
		var timeSec float32
		fmt.Sscanf(line, "size=%dkB time=%d:%d:%f bitrate=%fkbits/s speed=%fx", &s.Status.Size, &timeHour, &timeMin, &timeSec, &s.Status.Bitrate, &s.Status.Speed)

		s.Status.Time, _ = time.ParseDuration(fmt.Sprintf("%dh%dm%.2fs", timeHour, timeMin, timeSec))
	}
}

func (s *encodeSession) readStdOut(stdOut io.Reader) {
	decoder := ogg.NewPacketDecoder(ogg.NewDecoder(stdOut))

	// 最初の2つは ogg opus metadata
	skipPacket := 2
	for {
		packet, _, err := decoder.Decode()
		if skipPacket > 0 {
			skipPacket--
			continue
		}

		if err != nil {
			if err != io.EOF {
				s.done <- err
			}
			return
		}

		err = s.writeOpus(packet)
		if err != nil {
			s.done <- err
			return
		}
	}
}

func (s *encodeSession) writeOpus(opus []byte) error {
	var dcaBuf bytes.Buffer

	err := binary.Write(&dcaBuf, binary.LittleEndian, int16(len(opus)))
	if err != nil {
		return err
	}

	_, err = dcaBuf.Write(opus)
	if err != nil {
		return err
	}

	s.flames <- dcaBuf.Bytes()
	return nil
}

func (s *encodeSession) stream() {
	for {
		err := s.readNext()
		if err != nil {
			if s.done != nil {
				go func() {
					s.done <- err
				}()
				break
			}
		}
	}
}

func (s *encodeSession) readNext() error {
	opus, err := s.opusFlame()
	if err != nil {
		return err
	}

	// vc write timeout
	timeout := time.NewTicker(1 * time.Second)

	select {
	case <-timeout.C:
		return fmt.Errorf("voice connection is timed out")
	case s.vc.OpusSend <- opus:
		timeout.Stop()
	}
	return nil
}

func (s *encodeSession) opusFlame() (frame []byte, err error) {
	f := <-s.flames
	if f == nil {
		return nil, io.EOF
	}
	if len(f) < 2 {
		return nil, fmt.Errorf("flame data loss / bad flame data")
	}

	return f[2:], nil
}

func (s *encodeSession) Stop() {
	if s.process == nil {
		return
	}
	s.process.Kill()
}

func (s *encodeSession) Cleanup() {
	s.Stop()

	if s.isFlamesClosed {
		return
	}
	close(s.flames)
	s.isFlamesClosed = true
}
