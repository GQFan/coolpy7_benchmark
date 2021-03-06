package clientsession

import (
	"sync"

	"coolpy7_benchmark/src/packet"
)

// PacketStore is a goroutine safe packet store.
type PacketStore struct {
	packets map[packet.ID]packet.GenericPacket
	mutex   sync.RWMutex
}

// NewPacketStore returns a new PacketStore.
func NewPacketStore() *PacketStore {
	return &PacketStore{
		packets: make(map[packet.ID]packet.GenericPacket),
	}
}

// Save will store a packet in the store. An eventual existing packet with the
// same id gets quietly overwritten.
func (s *PacketStore) Save(pkt packet.GenericPacket) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id, ok := packet.GetID(pkt)
	if ok {
		s.packets[id] = pkt
	}
}

// Lookup will retrieve a packet from the store.
func (s *PacketStore) Lookup(id packet.ID) packet.GenericPacket {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.packets[id]
}

// Delete will remove a packet from the store.
func (s *PacketStore) Delete(id packet.ID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.packets, id)
}

// All will return all packets currently saved in the store.
func (s *PacketStore) All() []packet.GenericPacket {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var all []packet.GenericPacket

	for _, pkt := range s.packets {
		all = append(all, pkt)
	}

	return all
}

// Reset will reset the store.
func (s *PacketStore) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.packets = make(map[packet.ID]packet.GenericPacket)
}
