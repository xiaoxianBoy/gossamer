// Copyright 2019 ChainSafe Systems (ON) Corp.
// This file is part of gossamer.
//
// The gossamer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gossamer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the gossamer library. If not, see <http://www.gnu.org/licenses/>.

package p2p

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	mrand "math/rand"
	"path"
	"path/filepath"

	log "github.com/ChainSafe/log15"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	ma "github.com/multiformats/go-multiaddr"
)

const KeyFile = "node.key"

// Config is used to configure a p2p service
type Config struct {
	// Peers used for bootstrapping
	BootstrapNodes []string
	// Listening port
	Port uint32
	// If 0, random host ID will be generated; If non-0, deterministic ID will be produced, keys will not be loaded from data dir
	RandSeed int64
	// Disable bootstrapping altogether. BootstrapNodes has no effect over this.
	NoBootstrap bool
	// Disables MDNS discovery
	NoMdns bool
	// Global data directory
	DataDir string
	// Identity key for node
	privateKey crypto.PrivKey
}

func (c *Config) buildOpts() ([]libp2p.Option, error) {
	ip := "0.0.0.0"

	if c.RandSeed == 0 {
		err := c.setupPrivKey()
		if err != nil {
			return nil, err
		}
	} else {
		key, err := generateKey(c.RandSeed, c.DataDir)
		if err != nil {
			return nil, err
		}
		c.privateKey = key
	}

	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ip, c.Port))
	if err != nil {
		return nil, err
	}

	connMgr := ConnManager{}

	return []libp2p.Option{
		libp2p.ListenAddrs(addr),
		libp2p.DisableRelay(),
		libp2p.Identity(c.privateKey),
		libp2p.NATPortMap(),
		libp2p.Ping(true),
		libp2p.ConnectionManager(connMgr),
	}, nil
}

// setupPrivKey will attempt to load the nodes private key, if that fails it will create one
func (c *Config) setupPrivKey() error {
	// If key exists, load it
	key, err := tryLoadPrivKey(c.DataDir)
	if err != nil {
		return err
	}
	// Otherwise, create a key
	if key == nil {
		log.Debug("No existing p2p key, generating a new one")
		key, err = generateKey(c.RandSeed, c.DataDir)
		if err != nil {
			return err
		}
	}

	c.privateKey = key
	return nil
}

// tryLoadPrivkey will attempt to load the private key from the provided path
func tryLoadPrivKey(fp string) (crypto.PrivKey, error) {
	keyData, err := ioutil.ReadFile(path.Join(filepath.Clean(fp), KeyFile))
	if err != nil {
		return nil, nil
	}

	return crypto.UnmarshalEd25519PrivateKey(keyData)
}

// generateKey generates an ed25519 private key and writes it to the data directory
// If the seed is zero, we use real cryptographic randomness. Otherwise, we use a
// deterministic randomness source to make generated keys stay the same
// across multiple runs
func generateKey(seed int64, fp string) (crypto.PrivKey, error) {

	var r io.Reader
	if seed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(seed))
	}

	// Generate a key pair for this host. We will use it at least
	// to obtain a valid host ID.
	_, priv, err := ed25519.GenerateKey(r)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path.Join(filepath.Clean(fp), KeyFile), priv, 0600)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalEd25519PrivateKey(priv)
}