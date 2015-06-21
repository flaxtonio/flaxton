import json
import requests
import time

def AddContainerCall(container_id, daemon):
	print(daemon)
	cont = []
	cont.append(container_id)
	cont.append("aaaa")
	url = 'http://container.flaxton.io/api/task/add'
	send_data = {'task_type' : 'pause_container',
		'daemon' : daemon,
		'data': cont
	}
	r = requests.post(url, data=send_data)
	print(r.text)

last_container = []

while True:
	r = requests.get('http://container.flaxton.io/daemon-state')
	data = json.loads(r.text)
	for daemon, info in data.items():
		if "containers" in info["data"]:
			if len(info["data"]["containers"].keys()) == 2 and len(last_container) != 0:
				AddContainerCall(last_container, daemon)
				last_container = []
				time.sleep(10)
			else:
				first = list(info["data"]["containers"].keys())[0]
				last_container = info["data"]["containers"][first]["inspect"]["Config"]["Hostname"]
	time.sleep(2)