package protocol

type ProtocolVersion struct {
    Version    string
    Protocol   int
    DataVersion int
    Features   VersionFeatures
}

type VersionFeatures struct {
    SupportsNewBlocks    bool
    SupportsNewEntities  bool
    SupportsNewItems     bool
    CompressionThreshold int
    RakNetVersion        byte
}

var Versions = map[string]ProtocolVersion{
    "1.21.10": {
        Version:      "1.21.10",
        Protocol:     671,
        DataVersion:  39597969,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.11": {
        Version:      "1.21.11",
        Protocol:     672,
        DataVersion:  39607368,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.12": {
        Version:      "1.21.12",
        Protocol:     673,
        DataVersion:  39616767,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.13": {
        Version:      "1.21.13",
        Protocol:     674,
        DataVersion:  39626166,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.14": {
        Version:      "1.21.14",
        Protocol:     675,
        DataVersion:  39635565,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.15": {
        Version:      "1.21.15",
        Protocol:     676,
        DataVersion:  39644964,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.20": {
        Version:      "1.21.20",
        Protocol:     677,
        DataVersion:  39654363,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.21": {
        Version:      "1.21.21",
        Protocol:     678,
        DataVersion:  39663762,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.22": {
        Version:      "1.21.22",
        Protocol:     679,
        DataVersion:  39673161,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.30": {
        Version:      "1.21.30",
        Protocol:     680,
        DataVersion:  39682560,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.31": {
        Version:      "1.21.31",
        Protocol:     681,
        DataVersion:  39691959,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.32": {
        Version:      "1.21.32",
        Protocol:     682,
        DataVersion:  39701358,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.33": {
        Version:      "1.21.33",
        Protocol:     683,
        DataVersion:  39710757,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.40": {
        Version:      "1.21.40",
        Protocol:     684,
        DataVersion:  39720156,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.41": {
        Version:      "1.21.41",
        Protocol:     685,
        DataVersion:  39729555,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.42": {
        Version:      "1.21.42",
        Protocol:     686,
        DataVersion:  39738954,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.43": {
        Version:      "1.21.43",
        Protocol:     687,
        DataVersion:  39748353,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.50": {
        Version:      "1.21.50",
        Protocol:     688,
        DataVersion:  39757752,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.51": {
        Version:      "1.21.51",
        Protocol:     689,
        DataVersion:  39767151,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.52": {
        Version:      "1.21.52",
        Protocol:     690,
        DataVersion:  39776550,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.53": {
        Version:      "1.21.53",
        Protocol:     691,
        DataVersion:  39785949,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.60": {
        Version:      "1.21.60",
        Protocol:     692,
        DataVersion:  39795348,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.61": {
        Version:      "1.21.61",
        Protocol:     693,
        DataVersion:  39804747,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.62": {
        Version:      "1.21.62",
        Protocol:     694,
        DataVersion:  39814146,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.63": {
        Version:      "1.21.63",
        Protocol:     695,
        DataVersion:  39823545,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.70": {
        Version:      "1.21.70",
        Protocol:     696,
        DataVersion:  39832944,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.71": {
        Version:      "1.21.71",
        Protocol:     697,
        DataVersion:  39842343,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.72": {
        Version:      "1.21.72",
        Protocol:     698,
        DataVersion:  39851742,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.73": {
        Version:      "1.21.73",
        Protocol:     699,
        DataVersion:  39861141,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.80": {
        Version:      "1.21.80",
        Protocol:     700,
        DataVersion:  39870540,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.81": {
        Version:      "1.21.81",
        Protocol:     701,
        DataVersion:  39879939,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.82": {
        Version:      "1.21.82",
        Protocol:     702,
        DataVersion:  39889338,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.83": {
        Version:      "1.21.83",
        Protocol:     703,
        DataVersion:  39898737,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.90": {
        Version:      "1.21.90",
        Protocol:     704,
        DataVersion:  39908136,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.91": {
        Version:      "1.21.91",
        Protocol:     705,
        DataVersion:  39917535,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.92": {
        Version:      "1.21.92",
        Protocol:     706,
        DataVersion:  39926934,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.93": {
        Version:      "1.21.93",
        Protocol:     707,
        DataVersion:  39936333,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.100": {
        Version:      "1.21.100",
        Protocol:     708,
        DataVersion:  39945732,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.101": {
        Version:      "1.21.101",
        Protocol:     709,
        DataVersion:  39955131,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.102": {
        Version:      "1.21.102",
        Protocol:     710,
        DataVersion:  39964530,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.103": {
        Version:      "1.21.103",
        Protocol:     711,
        DataVersion:  39973929,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.110": {
        Version:      "1.21.110",
        Protocol:     712,
        DataVersion:  39983328,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.111": {
        Version:      "1.21.111",
        Protocol:     713,
        DataVersion:  39992727,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.112": {
        Version:      "1.21.112",
        Protocol:     714,
        DataVersion:  40002126,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.113": {
        Version:      "1.21.113",
        Protocol:     715,
        DataVersion:  40011525,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.120": {
        Version:      "1.21.120",
        Protocol:     716,
        DataVersion:  40020924,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.121": {
        Version:      "1.21.121",
        Protocol:     717,
        DataVersion:  40030323,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.122": {
        Version:      "1.21.122",
        Protocol:     718,
        DataVersion:  40039722,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
    "1.21.123": {
        Version:      "1.21.123",
        Protocol:     719,
        DataVersion:  40049121,
        Features: VersionFeatures{
            SupportsNewBlocks:   true,
            SupportsNewEntities: true,
            SupportsNewItems:    true,
            CompressionThreshold: 512,
            RakNetVersion:       11,
        },
    },
}

// Helper functions
func GetVersion(version string) (ProtocolVersion, bool) {
    v, exists := Versions[version]
    return v, exists
}

func GetVersionByProtocol(protocol int) (ProtocolVersion, bool) {
    for _, v := range Versions {
        if v.Protocol == protocol {
            return v, true
        }
    }
    return ProtocolVersion{}, false
}

func GetLatestVersion() ProtocolVersion {
    // Return the latest version (1.21.123)
    return Versions["1.21.123"]
}

func IsVersionSupported(version string) bool {
    _, exists := Versions[version]
    return exists
}

func GetSupportedVersions() []string {
    versions := make([]string, 0, len(Versions))
    for v := range Versions {
        versions = append(versions, v)
    }
    return versions
}