import re

import requests
from bs4 import BeautifulSoup

headers = {
    'Cookie': 'GET FROM UBER.COM',
    'x-ajax-replace': 'true'
}

cost_strings = []
for i in range(1, 100):
    print(i)
    resp = requests.get('https://riders.uber.com/trips?page='+str(i), headers=headers)
    soup = BeautifulSoup(resp.text, 'html.parser')
    trips = soup.findAll("tr", {"data-target" : re.compile('#trip.*')})
    if len(trips) == 0:
        break
    for trip in trips:
        cost_tag = trip.find_all('td')[3]
        spans = cost_tag.find_all('span')
        if len(spans) > 0:
            spans[0].extract()
        cost_string = cost_tag.string
        if cost_string is None or cost_string == 'Canceled':
            continue
        cost_strings.append(cost_string)

rates = requests.get('http://api.fixer.io/latest?base=USD').json()
sums = 0
for s in cost_strings:
    m = re.match("\$(\d+\.\d+)", s)
    if m:
        sums += float(m.group(1))
        continue
    m = re.match("CA\$(\d+\.\d+)", s)
    if m:
        sums += (float(m.group(1)) / rates['rates']['CAD'])
        continue
    m = re.match("THB(\d+\.\d+)", s)
    if m:
        sums += (float(m.group(1)) / rates['rates']['THB'])
        continue
print(sums)
