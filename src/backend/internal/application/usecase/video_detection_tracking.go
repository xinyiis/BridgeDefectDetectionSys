package usecase

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
)

type defectTrackStatus string

const (
	trackCandidate defectTrackStatus = "candidate"
	trackConfirmed defectTrackStatus = "confirmed"
	trackClosed    defectTrackStatus = "closed"
)

type activeDefectTrack struct {
	TrackID           string
	TaskID            string
	BridgeID          uint
	DefectType        string
	Status            defectTrackStatus
	FirstSeenFrame    int
	LastSeenFrame     int
	Hits              int
	Misses            int
	LastBBox          dto.VideoBBox
	BestBBox          dto.VideoBBox
	BestConfidence    float64
	AvgConfidence     float64
	BestFramePath     string
	BestResultPath    string
	DefectID          *uint
	FirstSeenAt       time.Time
	LastSeenAt        time.Time
	FinalLength       float64
	FinalWidth        float64
	FinalArea         float64
	MeasurementSource string
	MeasurementStatus string
}

func (uc *VideoDetectionUseCase) loadTracks(taskID string) []*activeDefectTrack {
	if value, ok := uc.tracks.Load(taskID); ok {
		if tracks, ok := value.([]*activeDefectTrack); ok {
			return tracks
		}
	}
	return []*activeDefectTrack{}
}

func (uc *VideoDetectionUseCase) storeTracks(taskID string, tracks []*activeDefectTrack) {
	uc.tracks.Store(taskID, tracks)
}

func (uc *VideoDetectionUseCase) upsertTrack(
	task *model.VideoAnalysisTask,
	tracks []*activeDefectTrack,
	current dto.VideoFrameDefect,
	confidence float64,
	frameNo int,
	observedAt time.Time,
) (*activeDefectTrack, bool) {
	var bestTrack *activeDefectTrack
	bestScore := 0.0

	for _, track := range tracks {
		if track.Status == trackClosed {
			continue
		}
		if track.DefectType != current.DefectType {
			continue
		}
		if frameNo-track.LastSeenFrame > 3 {
			continue
		}

		score, matched := trackMatchScore(track.LastBBox, current.BBox)
		if matched && score > bestScore {
			bestScore = score
			bestTrack = track
		}
	}

	if bestTrack == nil {
		track := &activeDefectTrack{
			TrackID:           "track_" + uuid.NewString(),
			TaskID:            task.TaskID,
			BridgeID:          task.BridgeID,
			DefectType:        current.DefectType,
			Status:            trackCandidate,
			FirstSeenFrame:    frameNo,
			LastSeenFrame:     frameNo,
			Hits:              1,
			LastBBox:          current.BBox,
			BestBBox:          current.BBox,
			BestConfidence:    confidence,
			AvgConfidence:     confidence,
			FirstSeenAt:       observedAt,
			LastSeenAt:        observedAt,
			FinalLength:       current.Length,
			FinalWidth:        current.Width,
			FinalArea:         current.Area,
			MeasurementSource: current.MeasurementSource,
			MeasurementStatus: current.MeasurementStatus,
		}
		return track, true
	}

	bestTrack.Hits++
	bestTrack.Misses = 0
	bestTrack.LastSeenFrame = frameNo
	bestTrack.LastBBox = current.BBox
	bestTrack.LastSeenAt = observedAt
	bestTrack.AvgConfidence = ((bestTrack.AvgConfidence * float64(bestTrack.Hits-1)) + confidence) / float64(bestTrack.Hits)

	becameBest := bestTrack.BestFramePath == "" || confidence > bestTrack.BestConfidence
	if becameBest {
		bestTrack.BestBBox = current.BBox
		bestTrack.BestConfidence = confidence
		bestTrack.FinalLength = current.Length
		bestTrack.FinalWidth = current.Width
		bestTrack.FinalArea = current.Area
		bestTrack.MeasurementSource = current.MeasurementSource
		bestTrack.MeasurementStatus = current.MeasurementStatus
	}

	return bestTrack, becameBest
}

func (uc *VideoDetectionUseCase) closeMissedTracks(tracks []*activeDefectTrack, matchedTrackIDs map[string]struct{}) {
	for _, track := range tracks {
		if track.Status == trackClosed {
			continue
		}
		if _, matched := matchedTrackIDs[track.TrackID]; matched {
			continue
		}

		track.Misses++
		if track.Misses > 3 {
			track.Status = trackClosed
		}
	}
}

func (uc *VideoDetectionUseCase) shouldConfirmTrack(track *activeDefectTrack) bool {
	return track.Hits >= 3 && track.AvgConfidence >= 0.7
}

func trackMatchScore(previous dto.VideoBBox, current dto.VideoBBox) (float64, bool) {
	iou := bboxIOU(previous, current)
	centerDistance := normalizedCenterDistance(previous, current)

	if iou > 0.30 || centerDistance < 0.08 {
		centerSimilarity := math.Max(0, 1-centerDistance)
		return math.Max(iou, centerSimilarity), true
	}

	return 0, false
}

func bboxIOU(a dto.VideoBBox, b dto.VideoBBox) float64 {
	ax2 := a.X + a.Width
	ay2 := a.Y + a.Height
	bx2 := b.X + b.Width
	by2 := b.Y + b.Height

	interLeft := maxInt(a.X, b.X)
	interTop := maxInt(a.Y, b.Y)
	interRight := minInt(ax2, bx2)
	interBottom := minInt(ay2, by2)

	interWidth := interRight - interLeft
	interHeight := interBottom - interTop
	if interWidth <= 0 || interHeight <= 0 {
		return 0
	}

	interArea := float64(interWidth * interHeight)
	unionArea := float64(a.Width*a.Height+b.Width*b.Height) - interArea
	if unionArea <= 0 {
		return 0
	}

	return interArea / unionArea
}

func normalizedCenterDistance(a dto.VideoBBox, b dto.VideoBBox) float64 {
	ax := float64(a.X) + float64(a.Width)/2
	ay := float64(a.Y) + float64(a.Height)/2
	bx := float64(b.X) + float64(b.Width)/2
	by := float64(b.Y) + float64(b.Height)/2

	dx := ax - bx
	dy := ay - by
	distance := math.Sqrt(dx*dx + dy*dy)
	normalizer := math.Max(
		math.Max(float64(a.Width), float64(a.Height)),
		math.Max(float64(b.Width), float64(b.Height)),
	)
	if normalizer == 0 {
		return 1
	}

	return distance / normalizer
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
