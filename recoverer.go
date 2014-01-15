package grass

import (
	"encoding/json"

	"github.com/crbrox/store"
)

type Recoverer struct {
	SendTo        chan<- *Petition
	PetitionStore store.Interface
}

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
		r.SendTo <- pet
	}
	return err
}
