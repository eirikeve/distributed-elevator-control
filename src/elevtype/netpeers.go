package elevtype

/*PeerUpdate is a struct representing the state of the Peers we have in a P2P network
 *It consists of the currently active peers, along with lost and new peers
 */
type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}
