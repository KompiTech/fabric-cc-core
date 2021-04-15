package engine

import (
	"fmt"
	"strconv"

	. "github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

// Changelog stores history of all schema changes in Registry
type Changelog struct {
	ctx  ContextInterface
	head int // head is the latest number of changelog item
}

func NewChangelog(ctx ContextInterface) (*Changelog, error) {
	// count number of existing changelog items to get head
	iterator, err := ctx.Stub().GetStateByPartialCompositeKey(ChangelogItemPrefix, []string{})
	if err != nil {
		return &Changelog{}, errors.Wrap(err, "ctx.Stub().GetStateByPartialCompositeKey() failed")
	}
	defer func() { _ = iterator.Close() }()

	numberMax := 0

	for iterator.HasNext() {
		item, err := iterator.Next()
		if err != nil {
			return &Changelog{}, errors.Wrap(err, "iterator.Next() failed")
		}

		_, keyElems, err := ctx.Stub().SplitCompositeKey(item.GetKey())
		if err != nil {
			return &Changelog{}, errors.Wrap(err, "ctx.Stub().SplitCompositeKey() failed")
		}

		if len(keyElems) != 1 {
			return &Changelog{}, fmt.Errorf("existing changelog item key: %s does not have expected format", item.GetKey())
		}

		// parse version as int from last elem of composite key
		number, err := strconv.Atoi(keyElems[0])
		if err != nil {
			return &Changelog{}, errors.Wrap(err, "strconv.Atoi() failed")
		}

		if number > numberMax {
			numberMax = number
		}
	}

	return &Changelog{ctx, numberMax}, nil
}

func (c *Changelog) Create(ci ChangelogItem) error {
	c.head = c.head + 1
	number := strconv.Itoa(c.head)

	key, err := c.ctx.Stub().CreateCompositeKey(ChangelogItemPrefix, []string{number})
	if err != nil {
		return errors.Wrap(err, "ctx.Stub().CreateCompositeKey() failed")
	}

	if err := putRmapToState(c.ctx, key, true, ci.Rmap()); err != nil {
		return errors.Wrap(err, "putRmapToState() failed")
	}

	return nil
}

func (c Changelog) Get(number int) (rmap.Rmap, error) {
	if number <= 0 {
		number = c.head
	}

	if number > c.head {
		return rmap.Rmap{}, fmt.Errorf("invalid changelog number: %d, head: %d", number, c.head)
	}

	key, err := c.ctx.Stub().CreateCompositeKey(ChangelogItemPrefix, []string{strconv.Itoa(number)})
	if err != nil {
		return rmap.Rmap{}, errors.Wrap(err, "ctx.Stub().CreateCompositeKey() failed")
	}

	ci, err := newRmapFromState(c.ctx, key, true)
	if err != nil {
		return rmap.Rmap{}, errors.Wrap(err, "newRmapFromState() failed")
	}

	return ci, nil
}

func (c Changelog) List() ([]rmap.Rmap, error) {
	output := make([]rmap.Rmap, 0, c.head)

	// iterate through number 1 - head
	for number := 1; number <= c.head; number += 1 {
		key, err := c.ctx.Stub().CreateCompositeKey(ChangelogItemPrefix, []string{strconv.Itoa(number)})
		if err != nil {
			return nil, errors.Wrap(err, "c.ctx.Stub().CreateCompositeKey() failed")
		}

		item, err := newRmapFromState(c.ctx, key, true)
		if err != nil {
			return nil, errors.Wrap(err, "newRmapFromState() failed")
		}

		output = append(output, item)
	}

	return output, nil
}
