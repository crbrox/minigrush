package minigrush

import (
	"encoding/json"

	"github.com/crbrox/store"
)

//Recoverer takes the petitions stored in PetitionStore and enqueues them again into SendTo.
type Recoverer struct {
	SendTo        chan<- *Petition
	PetitionStore store.Interface
}

//Recover gets all the petitions stored and sends them to a channel for processing by a consumer.
// It returns when all of them are re-enqueued or when an error happens. It should be run before starting
//a listener (with the same PetitionStore) or new petitions could be enqueued twice. Listeners with a different PetitionStore
//should not be a problem. A consumer can be started before with the same PetitionStore to avoid overflowing the queue.
func (r *Recoverer) Recover() error {
	ids, err := r.PetitionStore.List()
	if err != nil {
		return err
	}
	for _, id := range ids {
		data, err := r.PetitionStore.Get(id)
		if err != nil {
			return err
		}
		pet := &Petition{}
		err = json.Unmarshal(data, pet)
		if err != nil {
			return err
		}
		r.SendTo <- pet
	}
	return nil
}
