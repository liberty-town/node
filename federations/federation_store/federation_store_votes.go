package federation_store

import (
	"errors"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/store_data/polls"
	"liberty-town/node/federations/federation_store/store_data/polls/vote"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/store/store_db/store_db_interface"
	"liberty-town/node/store/store_utils"
)

func StoreVote(newVote *vote.Vote) (poll *polls.Poll, err error) {

	f := federation_serve.ServeFederation.Load()

	if f == nil || !f.Federation.Ownership.Address.Equals(newVote.FederationIdentity) {
		return nil, errors.New("not serving this federation")
	}

	if err = newVote.Validate(); err != nil {
		return
	}
	if err = newVote.ValidateSignatures(); err != nil {
		return
	}
	if !f.Federation.IsValidationAccepted(newVote.Validation) {
		return nil, errors.New("validation signature is not accepted")
	}

	err = f.Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		poll = &polls.Poll{
			FederationIdentity: f.Federation.Ownership.Address,
			Identity:           newVote.Identity,
		}

		b := tx.Get("polls:" + newVote.Identity.Encoded)
		if b != nil {
			if err := poll.Deserialize(advanced_buffers.NewBufferReader(b)); err != nil {
				return err
			}
		} else {
			poll.List = []*vote.Vote{}
			poll.Identity = newVote.Identity
			poll.FederationIdentity = newVote.FederationIdentity
		}

		if ok := poll.MergeVote(newVote); !ok {
			return errors.New("vote is not better")
		}

		tx.Put("polls:"+poll.Identity.Encoded, helpers.SerializeToBytes(poll))

		if err = store_utils.IncreaseCount("polls", poll.Identity.Encoded, poll.GetBetterScore(), tx); err != nil {
			return
		}

		if err = storeThreadScore(f.Federation, tx, poll.Identity, nil, false, poll); err != nil {
			return
		}

		return nil
	})

	return
}

func StorePoll(newPoll *polls.Poll) error {

	f := federation_serve.ServeFederation.Load()

	if f == nil || !f.Federation.Ownership.Address.Equals(newPoll.FederationIdentity) {
		return errors.New("not serving this federation")
	}

	if err := newPoll.Validate(); err != nil {
		return err
	}
	if err := newPoll.ValidateSignatures(); err != nil {
		return err
	}

	for i := range newPoll.List {
		if !f.Federation.IsValidationAccepted(newPoll.List[i].Validation) {
			return errors.New("validation signature is not accepted")
		}
	}

	return f.Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		oldPoll := &polls.Poll{
			FederationIdentity: f.Federation.Ownership.Address,
			Identity:           newPoll.Identity,
		}

		b := tx.Get("polls:" + newPoll.Identity.Encoded)
		if b != nil {
			if err := oldPoll.Deserialize(advanced_buffers.NewBufferReader(b)); err != nil {
				return err
			}
		} else {
			oldPoll.List = []*vote.Vote{}
			oldPoll.Identity = newPoll.Identity
			oldPoll.FederationIdentity = newPoll.FederationIdentity
		}

		ok := false
		for i := range newPoll.List {
			if ok2 := oldPoll.MergeVote(newPoll.List[i]); ok2 {
				ok = true
			}
		}
		if !ok {
			return errors.New("polls are not better")
		}

		tx.Put("polls:"+oldPoll.Identity.Encoded, helpers.SerializeToBytes(oldPoll))

		if err = store_utils.IncreaseCount("polls", oldPoll.Identity.Encoded, oldPoll.GetBetterScore(), tx); err != nil {
			return
		}

		if err = storeThreadScore(f.Federation, tx, oldPoll.Identity, nil, false, oldPoll); err != nil {
			return
		}

		return nil
	})
}
