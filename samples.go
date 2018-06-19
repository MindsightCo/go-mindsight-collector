package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const DEFAULT_CACHE_DEPTH = 100

type sampleCache struct {
	depth     int
	remaining int
	url       string
	samples   map[string]int
}

func newSampleCache(depth int, url string) *sampleCache {
	return &sampleCache{
		depth:     depth,
		remaining: depth,
		url:       url,
		samples:   make(map[string]int),
	}
}

func (c *sampleCache) sendSamples() error {
	payload, err := json.Marshal(c.samples)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("response status: %s, body: %s", resp.Status, string(body))
	}

	c.samples = make(map[string]int)
	return nil
}

func (c *sampleCache) recordSample(fn string) error {
	c.samples[fn] += 1
	c.remaining--

	if c.remaining <= 0 {
		if err := c.sendSamples(); err != nil {
			return err
		}
		c.remaining = c.depth
	}

	return nil
}
