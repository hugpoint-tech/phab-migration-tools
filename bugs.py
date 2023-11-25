import requests
import json

base_url = 'https://bugs.freebsd.org/bugzilla'
endpoint = '/rest/bug'

# Authenticate with Bugzilla
session = requests.Session()
session.post(base_url + '/rest/login', data={
    'login': 'ihor@antonovs.family',
    'password': r"""eavT<(W,?5G0:]_-KIv7U'yud""",
})

# Retrieve data in batches
batch_size = 1000
offset = 0
all_data = []
while True:
    response = session.get(base_url + endpoint, params={
        'limit': batch_size,
        'offset': offset,
    })
    if not response.ok:
        raise Exception('API call failed: {}'.format(response.content))
    data = json.loads(response.content)
    all_data.extend(data['bugs'])
    if len(data['bugs']) < batch_size:
        break
    offset += batch_size

# Save data to file or database
with open('bug_data.json', 'w') as f:
    json.dump(all_data, f)
