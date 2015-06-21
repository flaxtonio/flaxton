import json
import requests
import urllib2
import time

def AddContainerCall(container_id, daemon):
	url = 'http://container.flaxton.io/api/task/add'
	send_data = {'task_type' : 'start_container',
		'daemon' : daemon,
		'data': [container_id]
	}
	r = requests.post(url, data=send_data)

last_container = {}

while True:
	data = json.load(urllib2.urlopen('http://container.flaxton.io/daemon-state'))
	for daemon, info in data.items():
		print len(info["data"]["containers"].keys())
		if len(info["data"]["containers"].keys()) == 2:
			print last_container
			AddContainerCall(last_container, daemon)
		else:
			last_container = info["data"]["containers"].keys()[0]
	time.sleep(2)