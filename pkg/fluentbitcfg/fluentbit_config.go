package fluentbitcfg

import (
	"strings"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)


// Make fluent-bit.conf for DaemonSet
func MakeFluentbitConfig(template string, logs , procs, os_monits, outputs, kafkas *corev1.ConfigMapList, node_group string) map[string]string {
	ins := ""

	// Log inputs
	for _, log := range logs.Items {
		input := ""
		if log.Data["log_kind"] == "k8s_pod_log" {
			input = k8s_pod_log
		} else if log.Data["log_kind"] == "rke_container_log" {
			input = rke_container_log
		} else if log.Data["log_kind"] == "syslog" {
			input = syslog
		} else {
			continue
		}
		input = strings.Replace(input, "@PATH", log.Data["path"], 1)
		input = strings.Replace(input, "@TAG", log.Data["tag"], 2)
		ins += input
	}

	// Proccess monitorings
	for _, proc := range procs.Items {
		if proc.Data["node_group"] != node_group {
			continue
		}
		proc_names := strings.Split(proc.Data["proc_names"],",")
		for _, proc_name := range proc_names {
			tag := strings.Replace(proc.Data["tag"], "*", proc_name, 1)
			input := proc_monitoring
			input = strings.Replace(input, "@TAG", tag, 1)
			input = strings.Replace(input, "@PROC_NAME", proc_name, 1)
			input = strings.Replace(input, "@INTERVAL", proc.Data["interval_sec"], 1)
			ins += input
		}
	}

	// OS Monitorings
	for _, monit := range os_monits.Items {
		// cpu
		if monit.Data["cpu_tag"] != "" {
			input := os_cpu
			input = strings.Replace(input, "@INTERVAL", monit.Data["cpu_interval_sec"], 1)
			input = strings.Replace(input, "@TAG", monit.Data["cpu_tag"], 1)
			ins += input
		}
		// memory
		if monit.Data["memory_tag"] != "" {
			input := os_memory
			input = strings.Replace(input, "@INTERVAL", monit.Data["memory_interval_sec"], 1)
			input = strings.Replace(input, "@TAG", monit.Data["memory_tag"], 1)
			ins += input
		}
		// disk io
		if monit.Data["io_tag"] != "" {
			input := os_io
			input = strings.Replace(input, "@INTERVAL", monit.Data["io_interval_sec"], 1)
			input = strings.Replace(input, "@NAME", monit.Data["io_diskname"], 1)
			input = strings.Replace(input, "@TAG", monit.Data["io_tag"], 1)
			ins += input
		}
		// filesystem usage
		if monit.Data["filesystem_tag"] != "" {
			input := os_filesystem
			input = strings.Replace(input, "@INTERVAL", monit.Data["filesystem_interval_sec"], 1)
			input = strings.Replace(input, "@DIR", monit.Data["filesystem_df_dir"], 1)
			input = strings.Replace(input, "@TAG", monit.Data["filesystem_tag"], 1)
			ins += input
		}
	}

	// Outputs
	outs := ""
	for _, out := range outputs.Items {
		output := es_output
		output = strings.Replace(output, "@MATCH", out.Data["match"], 1)
		output = strings.Replace(output, "@HOST", out.Data["host"], 1)
		output = strings.Replace(output, "@PORT", out.Data["port"], 1)
		output = strings.Replace(output, "@PREFIX", out.Data["index_prefix"], 1)
		outs += output
	}

	// Kafka outputs
	for _, out := range kafkas.Items {
		output := kafka_output
		output = strings.Replace(output, "@MATCH", out.Data["match"], 1)
		output = strings.Replace(output, "@BROKERS", out.Data["brokers"], 1)
		output = strings.Replace(output, "@TIMESTAMP_FORMAT", out.Data["timestamp_format"], 1)
		output = strings.Replace(output, "@TOPICS", out.Data["topics"], 1)
		rdkafka_options := ""
		if out.Data["rdkafka_options"] != "" {
			options := strings.Split(out.Data["rdkafka_options"],",")
			for _, option := range options {
				kv := strings.Split(option,"=")
				if len(kv) == 2 {
					rdkafka_options += fmt.Sprintf("    rdkafka.%s  %s\n", kv[0], kv[1])
				}
			}
		}
		output = strings.Replace(output, "@RDKAFKA_OPTIONS", rdkafka_options, 1)
		outs += output
	}

	config := template
	config = strings.Replace(config, "@INPUTS", ins, 1)
	config = strings.Replace(config, "@OUTPUTS", outs, 1)

	return map[string]string{"fluent-bit.conf":config}
}

// Make fluent-bit.conf for DaemonSet
func MakeFluentbitMetricsConfig(template string, metrics, apps, outputs, kafkas *corev1.ConfigMapList) map[string]string {

	ins := ""

	// K8s metrics inputs
	for _, m := range metrics.Items {
		input := ""
		if m.Data["metric_kind"] == "pod" {
			input = pod_metrics
		} else if m.Data["metric_kind"] == "node" {
			input = node_metrics
		} else {
			continue
		}
		input = strings.Replace(input, "@INTERVAL", m.Data["interval_sec"], 2)
		input = strings.Replace(input, "@TAG", m.Data["tag"], 2)
		ins += input
	}

	// K8s apps status inputs
	for _, app := range apps.Items {
		app_kinds := strings.Split(app.Data["app_kinds"],",")
		for _, kind := range app_kinds {
			input := ""
			switch kind {
			case "deployments":
				input = deployment_status
			case "daemonsets":
				input = daemonset_status
			case "statefulsets":
				input = statefulset_status
			default:
				continue
			}
			tag := strings.Replace(app.Data["tag"], "*", kind, 1)
			input = strings.Replace(input, "@TAG", tag, 1)
			input = strings.Replace(input, "@INTERVAL", app.Data["interval_sec"], 1)
			ins += input
		}
	}

	// Outputs
	outs := ""
	for _, out := range outputs.Items {
		output := es_output
		output = strings.Replace(output, "@MATCH", out.Data["match"], 1)
		output = strings.Replace(output, "@HOST", out.Data["host"], 1)
		output = strings.Replace(output, "@PORT", out.Data["port"], 1)
		output = strings.Replace(output, "@PREFIX", out.Data["index_prefix"], 1)
		outs += output
	}

	// Kafka outputs
	for _, out := range kafkas.Items {
		output := kafka_output
		output = strings.Replace(output, "@MATCH", out.Data["match"], 1)
		output = strings.Replace(output, "@BROKERS", out.Data["brokers"], 1)
		output = strings.Replace(output, "@TIMESTAMP_FORMAT", out.Data["timestamp_format"], 1)
		output = strings.Replace(output, "@TOPICS", out.Data["topics"], 1)
		rdkafka_options := ""
		if out.Data["rdkafka_options"] != "" {
			options := strings.Split(out.Data["rdkafka_options"],",")
			for _, option := range options {
				kv := strings.Split(option,"=")
				if len(kv) == 2 {
					rdkafka_options += fmt.Sprintf("    rdkafka.%s  %s\n", kv[0], kv[1])
				}
			}
		}
		output = strings.Replace(output, "@RDKAFKA_OPTIONS", rdkafka_options, 1)
		outs += output
	}

	config := template
	config = strings.Replace(config, "@INPUTS", ins, 1)
	config = strings.Replace(config, "@OUTPUTS", outs, 1)

	return map[string]string{"fluent-bit.conf":config}
}
