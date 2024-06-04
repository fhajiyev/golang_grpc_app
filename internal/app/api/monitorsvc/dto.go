package monitorsvc

// ESHealthRes type definition
type ESHealthRes []*ESHealth

// ESStatus type definition
type ESStatus string

// ESStatusOK status definition for elastic search's ok
const ESStatusOK = "green"

// ESHealth struct definition
type ESHealth struct {
	Cluster string   `json:"cluster"`
	Status  ESStatus `json:"status"`
}

// {
//     "epoch": "1575285378",
//     "timestamp": "11:16:18",
//     "cluster": "prod-buzzscreen-es-20170924",
//     "status": "green",
//     "node.total": "12",
//     "node.data": "9",
//     "shards": "59",
//     "pri": "19",
//     "relo": "0",
//     "init": "0",
//     "unassign": "0",
//     "pending_tasks": "0",
//     "max_task_wait_time": "-",
//     "active_shards_percent": "100.0%"
//  }
