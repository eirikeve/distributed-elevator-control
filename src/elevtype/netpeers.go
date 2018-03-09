package elevtype

/*PeerUpdate is a struct representing the state of the Peers we have in a P2P network
 *It consists of the currently active peers, along with the lost peers
 */
type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

/*UpdatePeers updates the Peers and Lost of p by checking the Peers of Update.
 * @arg p: The peer overview we want to update
 * @arg update: The update we want to apply to p
 */
func (p *PeerUpdate) UpdatePeers(update PeerUpdate) {
	newPeers := make([]string, len(update.Peers))
	copy(newPeers, update.Peers)

	p.addNewPeersAndLostPeersToLost(newPeers)
	p.removeCurrentPeersFromLost()
}

/*addNewPeersAndLostPeersToLost compares the entries of newPeers with those of p.Peers
 * The peers in p.Peers not found in newPeers, are added to p.Lost, since they are not currently active/reachable
 * p.Peers is then set to newPeers
 * @arg p: the peer overview we want to update
 * @arg newPeers: the currently active peers in the update we want to apply to p
 */
func (p *PeerUpdate) addNewPeersAndLostPeersToLost(newPeers []string) {
	// No need to check which peers to add if p.Peers is empty
	if len(p.Peers) == 0 {
		p.Peers = make([]string, len(newPeers))
		copy(p.Peers, newPeers)
	} else {
		// Iterate over all peers of p to determine which are not present in newPeers. These are added to p.Lost
		for i := 0; i < len(p.Peers); i++ {
			for j := 0; j < len(newPeers)+1; j++ {
				if j == len(newPeers) {
					// No Match => p.Peers[i] not found in newPeers
					newLost := make([]string, len(p.Lost)+1)
					copy(newLost, p.Lost)
					newLost[len(p.Lost)] = p.Peers[i]
					p.Lost = newLost
				} else {
					if p.Peers[i] == newPeers[j] {
						break // for j := 0 ..
					}
				}
			}
		}
	}
	p.Peers = newPeers

}

/*removeCurrentPeersFromLost checks if any entries in p.Lost are also in p.Peers, and thus not lost.
 * Those entries are then removed from Lost
 * @arg p: the peer overview to update
 */
func (p *PeerUpdate) removeCurrentPeersFromLost() {
	// Iterate over app Peers and Lost of p to find out which in Lost are now in Peers,
	for i := 0; i < len(p.Peers); i++ {
		for j := 0; j < len(p.Lost); j++ {
			if p.Peers[i] == p.Lost[j] {
				// Is not lost anymore, so remove it from the slice
				if j+1 < len(p.Lost) {
					p.Lost = append(p.Lost[:j], p.Lost[j+1:]...)
				} else {
					p.Lost = p.Lost[:j]
				}
			}
		}
	}
}
