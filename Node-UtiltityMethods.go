package AutoMPI

// GetBoundIPAddress get the bound IP address of the node
func (base *Node) GetBoundIPAddress() string {
	return base.localAddressString
}

// GetGUID get the GUID of the node
func (base *Node) GetGUID() string {
	return base.GUID
}
