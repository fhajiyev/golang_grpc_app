#!/usr/bin/env python
# coding: utf-8

# In[9]:


from __future__ import absolute_import, division, print_function, unicode_literals

"""
    * This notebook exists to be ran locally - tunneling must be enabled to successfully call the ES API endpoints.
    * You can use this to test out new scripts and ES queries (in json), and quickly understand the quality of articles
    being generated.
"""
script_file = 'v1_3_c.painless'
debug_es_query_file = 'debug_es_query.json'


# In[10]:


import requests
import urllib
import json
import pprint as pp
import datetime
get_ipython().run_line_magic('matplotlib', 'inline')
from IPython.core.display import display, HTML

# Oneline Painless script
f = open(script_file, 'r')
raw_script = f.read()
# inline_script = ''.join(raw_script.splitlines()) # \n lines are required for correct parsing of comments

# CTR
# f = open('v1_ctr.painless', 'r')
# raw_ctr = f.read()

# # Rec
# f = open('v1_rec.painless', 'r')
# raw_rec = f.read()

# Call ES API
f = open(debug_es_query_file, "r")
q = json.loads(f.read())


"""
Manually touches the json obj

"""
use_script = True
if use_script:
    inline = False
    if not inline:
        print('Loading external painless script...')
#         q['script_fields']['ctr']['script']['inline'] = raw_ctr
#         q['script_fields']['rec']['script']['inline'] = raw_rec
        
#         lines_ctr = raw_ctr.split('\n')
#         lines_rec = raw_rec.split('\n')
#         lines_all = lines_ctr[:-1] + lines_rec[:-1] + ['*'.join([lines_ctr[-1], lines_rec[-1]])]

        q['sort'][0]['_script']['script']['inline'] = raw_script

# Building personalization for device...

# This person is Male, born 1995, using SKT, using {HoneyScreen, OCB, etc.}, clicked on what kind of ads
# device = {'sex': 'M', 'year_of_birth': 1995, 'carrier': 'SKT', 'unit_id': '100000044'}

dev_index = 'buzzscreen-development-2017-10-24'
prod_index = 'buzzscreen-production-2018-05-21'
r = requests.post('http://localhost:39201/{}/content_campaign/_search?preference=_local&pretty=true'.format(prod_index),
                  json=q)
r=r.json()
r


# In[11]:


for rec in r['hits']['hits']:

    campaign_id = rec['_source']['id']
    title = rec['_source']['title']
    desc = rec['_source']['description']
    click_url = rec['_source']['click_url']
    img_url = rec['_source']['image']
    img_h, img_w = rec['_source']['image_height'], rec['_source']['image_width']
    pub_date = rec['_source']['published_at']
    category = rec['_source']['categories']

    if 'channel' not in rec['_source']:
        channel_logo_url = 'no channel img'
        channel_id = 'no id'
        channel_name = 'no channel name'
    else:
        channel_logo_url = rec['_source']['channel']['logo']
        channel_id = rec['_source']['channel']['id']
        channel_name = rec['_source']['channel']['name']

    if 'sort' not in rec:
        score = 'no score'
    else:
        score = rec['sort']


    # Calculate elapsed time, NOTE: a bit hacky
    now = datetime.datetime.utcnow()
    date_string = pub_date.split('+')[0].split('.')[0]
    past = datetime.datetime.strptime(date_string, "%Y-%m-%dT%H:%M:%S")
    minutes = round((now-past).total_seconds() / 60)

    if minutes == 0:
        elapsed_time, time_unit = '', 'now'
    elif minutes < 60:
        elapsed_time, time_unit = minutes, 'minutes ago'
    elif minutes < (60*24):
        elapsed_time, time_unit = round(minutes / 60), 'hours ago'
    else:
        elapsed_time, time_unit = round(minutes / (60*24)), 'days ago'


    display(HTML('<img src={} height={} width={}>'.format(img_url, img_h, img_w)))
    display(HTML('<a href={} target="_blank"><h1>{}</h1></a>'.format(click_url, title)))
    display(HTML('<img src={} height=64 width=64><div>{} - {} - {} - {} {} - {}</div>'.format(
                                                                                channel_logo_url,
                                                                                channel_name,
                                                                                category,
                                                                                score,
                                                                                elapsed_time,
                                                                                time_unit,
                                                                                campaign_id
                                                                            )))
#     display(HTML('<button onclick="(function(){{var x = document.getElementById(\'c_{}\');if (x.style.display === \'none\') {{x.style.display = \'block\';}} else {{x.style.display = \'none\';}})();return false;">Show JSON</button><div id="c_{}">{}</div>'.format(campaign_id, campaign_id, pp.pformat(rec))))
    pp.pprint(rec)
#     display(HTML('<p>{}</p>'.format(rec)))


# In[ ]:





# In[ ]:




