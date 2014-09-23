package datamodel

/******************************************
 ***** Summoner & snapshot structures *****
 ******************************************/

/**
 * An individual summoner, generated by joining several other data sources (see join-summoner executable).
 */
type SummonerRecord struct {
        SummonerId              uint32                                  `json:"id" bson:"_id"`
        SummonerName    string                                          `bson:"s"`
        LastUpdated             uint64                                          `bson:"l"`
        Daily                   map[string]*PlayerSnapshot      `bson:"d"`
        Weekly					map[string]*PlayerSnapshot		`bson:"w"`
        Monthly					map[string]*PlayerSnapshot		`bson:"m"`
        Metadata				SummonerMetadata				`bson:"e"`
}

type Metric interface {
}

type SimpleNumberMetric struct {
		Value	float64
}

/**
 * An individual snapshot for a summoner. Snapshots represent summaries of gameplay during a time range.
 */
type PlayerSnapshot struct {
        SummonerId              uint32                                          `bson:"s"`
        GamesList               []uint64                                        `bson:"g"`
        // TODO: add rank to this snapshot

        // The relevant gameplay statistics for the period covered by this snapshot
        Stats                   map[string]Metric                      `bson:"t"`

        // When this record was generated.
        CreationTimestamp       uint64                                  `bson:"c"`
}


/**
 * An individual statistic in a snapshot.
 */
/*
type PlayerStat struct {
        Name            string                                                  `bson:"n"`
        Absolute        float64                                         `bson:"a"`
        Normalized      uint32                                                  `bson:"o"`
}*/

type SummonerMetadata struct {
	SummonerId		uint32			`bson:"_id"`
	SummonerName	string			`bson:"n"`
}
