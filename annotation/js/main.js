$(document).ready(function() {
  //var top5chart;
  //var scoreChart;
  //var users = [];
  //var userIndex = -1;
  //var messages;
  //var messageIndex = -1;
  //var funcSkip;
  //var elasticsearch = $.es;
  //});

  function fetchMessage(user_id) {
    client.search({
      index: 'demo_twitter',
      type: 'result',
      body: {
        size: 10000,
        sort: [
          {
            relevantScore: {
              order: 'desc'
            }
          }
        ],
        query: {
          match: {
            profileid: user_id
          }
        }
      }
    }).then(function(resp) {
      processResponse(resp);
      showNextMessage();
    }, function(err) {
      if (err instanceof elasticsearch.errors.NoConnections) {
        console.log('Unable to connect to elasticsearch.');
      }
    });
  }

  function fetchAllUsers() {
    client.search({
      index: 'twitter_user_profile_v3',
      type: 'article',
      body: {
        size: 50,
        _source: "doc_desc",
        sort: [
          {
            profile_id: {
              order: 'asc'
            }
          }
        ]
      }
    }).then(function(resp) {
      for (var i = 0, len = resp.hits.total; i < len; i++) {
        var hit = resp.hits.hits[i];
        users.push({
          id: hit.sort[0],
          desc: hit._source.doc_desc
        });
      }
      nextUser();
    });
  }

  function nextUser() {
    funcSkip = nextUser;
    if (userIndex >= users.length - 1) return;
    var user = users[++userIndex];
    fetchMessage(user.id);
    updateUserProfile(user.id, user.desc);
  }

  function prevUser() {
    funcSkip = prevUser;
    if (userIndex == 0) return;
    var user = users[--userIndex];
    fetchMessage(user.id);
    updateUserProfile(user.id, user.desc);
  }

  function updateUserProfile(user_id, desc) {
    $('#user_id').text('Profile: ' + user_id);
    $('#user_profile').text(desc);
    updateTotal();
    updateChart();
  }

  function updateTotal() {
    // update total number
    client.get({
      index: 'demo_twitter',
      type: 'TotalNum',
      id: 1
    }).then(function(hit) {
      var totalnum = hit._source.totalnum;
      var mostRecentTime = new Date(hit._source.lastmatchtime + ' +0000');
      $('#total').html('<b>Last tweet matched </b> @ ' +
                       mostRecentTime.toLocaleDateString({timezone: 'America/New_York'}) + ' ' +
                       mostRecentTime.toLocaleTimeString({timezone: 'America/New_York'}) + '<br><b>Total tweets</b>: ' + totalnum);
    });
  }

  function processResponse(resp) {
    messages = [];
    messageIndex = -1;
    if (resp.hits.total == 0) {
      funcSkip();
    }
    for (var i = 0, len = resp.hits.total; i < len; i++) {
      var hit = resp.hits.hits[i];
      messages.push({
        user_id: hit._source.profileid,
        twitter: hit._source.text,
        date: new Date(hit._source.createtime + ' +0000'),
        delta: hit._source.differencetime,
        score: hit._source.relevantScore
      });
    }
  }

  function showNextMessage() {
    if (messageIndex >= messages.length - 1) return;
    var message = messages[++messageIndex];
    var point = scoreChart.series[0].points[0];
    point.update(Math.round(message.score * 100) / 100);
    $('#twits').html(' To user: ' + message.user_id +
                     '<br>Tweet on: ' + message.date.toLocaleDateString({timezone: 'America/New_York'}) +
                     '<br>    Time: ' + message.date.toLocaleTimeString({timezone: 'America/New_York'}) +
                       '<hr>' + message.twitter);
  }

  function showPrevMessage() {
    if (messageIndex == 0) return;
    var message = messages[--messageIndex];
    var point = scoreChart.series[0].points[0];
    point.update(Math.round(message.score * 100) / 100);
    $('#twits').html(' To user: ' + message.user_id +
                     '<br>Tweet on: ' + message.date.toLocaleDateString() +
                     '<br>    Time: ' + message.date.toLocaleTimeString() +
                       '<hr>' + message.twitter);
  }

  //initChart();
  //fetchAllUsers();
});
