package algoliautil

import (
	"fmt"
)

func IndexExistsInReplicas(replicas []string, indexName string, isVirtual bool) bool {
	replicaIndexName := getReplicaIndexName(indexName, isVirtual)
	for _, replica := range replicas {
		if replica == replicaIndexName {
			return true
		}
	}
	return false
}

func RemoveIndexFromReplicas(replicas []string, indexName string, isVirtual bool) []string {
	replicaIndexName := getReplicaIndexName(indexName, isVirtual)

	// Initialize as empty (non-nil) so the v4 SDK's MarshalJSON serializes an empty
	// `replicas` array when the last replica is removed; nil would be omitted entirely
	// and the API would leave the replicas list unchanged.
	newReplicas := []string{}
	for _, replica := range replicas {
		if replica == replicaIndexName {
			continue
		}
		newReplicas = append(newReplicas, replica)
	}
	return newReplicas
}

func getReplicaIndexName(indexName string, isVirtual bool) string {
	if isVirtual {
		return fmt.Sprintf("virtual(%s)", indexName)
	}
	return indexName
}
