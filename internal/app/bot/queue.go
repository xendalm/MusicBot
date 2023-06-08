package bot

import (
	"github.com/disgoorg/disgolink/v2/lavalink"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type QueueManager struct {
	queues map[string]*Queue
}

func (q *QueueManager) Get(guildID string) *Queue {
	queue, ok := q.queues[guildID]
	if !ok {
		queue = &Queue{
			Tracks:                   make([]lavalink.Track, 0),
			LastTrackToAddToPlaylist: nil,
		}
		q.queues[guildID] = queue
	}
	return queue
}

func (q *QueueManager) Delete(guildID string) {
	delete(q.queues, guildID)
}

type Queue struct {
	Tracks                   []lavalink.Track
	LastTrackToAddToPlaylist *lavalink.Track //needed to process the addition with the selection of a playlist in a separate interaction
}

func (q *Queue) Shuffle() {
	rand.Shuffle(len(q.Tracks), func(i, j int) {
		q.Tracks[i], q.Tracks[j] = q.Tracks[j], q.Tracks[i]
	})
}

func (q *Queue) Add(track ...lavalink.Track) {
	q.Tracks = append(q.Tracks, track...)
}

func (q *Queue) Next() (lavalink.Track, bool) {
	if len(q.Tracks) == 0 {
		return lavalink.Track{}, false
	}
	track := q.Tracks[0]
	q.Tracks = q.Tracks[1:]
	return track, true
}

func (q *Queue) Clear() {
	q.Tracks = make([]lavalink.Track, 0)
}

func (q *Queue) SaveTrackToAdd(track *lavalink.Track) {
	q.LastTrackToAddToPlaylist = track
}

func (q *Queue) GetSavedTrackToAdd() *lavalink.Track {
	return q.LastTrackToAddToPlaylist
}
