import json
import requests
import time

def AddContainerCall(container_id, daemon):
	url = 'http://container.flaxton.io/api/task/add'
	send_data = {'task_type' : 'pause_container',
		'daemon' : daemon,
		'data': [container_id]
	}
	r = requests.post(url, data=send_data)
	print(r.text)

last_container = {}

while True:
	r = requests.get('http://container.flaxton.io/daemon-state')
	data = json.loads(r.text)
	for daemon, info in data.items():
		print(len(info["data"]["containers"].keys()))
		if len(info["data"]["containers"].keys()) == 2 and len(last_container) != 0:
			print(last_container)
			AddContainerCall(last_container, daemon)
			exit(1)
		else:
			last_container = list(info["data"]["containers"].keys())[0]
	time.sleep(2)