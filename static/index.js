function openTab(evt, tabName) {
    var i, tabcontent, tablinks;
    if (evt)
    tabcontent = document.getElementsByClassName("tabcontent");
    for (i = 0; i < tabcontent.length; i++) {
      tabcontent[i].style.display = "none";
    }
    tablinks = document.getElementsByClassName("tablinks");
    for (i = 0; i < tablinks.length; i++) {
      tablinks[i].className = tablinks[i].className.replace(" active", "");
    }
    document.getElementById(tabName).style.display = "block";
    evt.currentTarget.className += " active";
  }
  function selectTab(tabName) {
    var i, tabcontent, tablinks;
    tabcontent = document.getElementsByClassName("tabcontent");
    for (i = 0; i < tabcontent.length; i++) {
      tabcontent[i].style.display = "none";
    }
    tablinks = document.getElementsByClassName("tablinks");
    for (i = 0; i < tablinks.length; i++) {
      tablinks[i].className = tablinks[i].className.replace(" active", "");
    }
    document.getElementById(tabName).style.display = "block";
    document.getElementById('searchTab').className += " active";
  }

  $(document).ready(function(){
        selectTab('Search')
   
        $('#submitSearch').click(function() {
            var term = $('#searchTerm').val();
            var searchTerm = {
                Term: term
            }
            document.getElementById("searchResults").innerHTML = "Searching...";                      

            $.ajax({
                url: 'http://localhost:7250/api/search',
                type: 'post',
                dataType: 'json',
                contentType: 'application/json',
                success: function (data) {
                    var txt = ""
                    if(data.Results) {
                        txt += "<table border='1'><tr><th>Url</th><th>Matches</th></tr>"
                        for (result of data.Results) {
                            txt += "<tr><td><a  href=" + result.URL + " target=_blank>" + result.Title + "</a></td><td>"+ result.Count +"</td></tr>"
                        }
                        txt += "</table>"
                    } else {
                        txt = "No results found"
                    }

                    document.getElementById("searchResults").innerHTML = txt;                      
                },
                error: function (err) {
                    document.getElementById("searchResults").innerHTML = "An error occurred during search";                      

                },
                data: JSON.stringify(searchTerm)
            });            
        });
        $('#submitCrawl').click(function() {
            var url = $('#indexUrl').val();
            var search = {
                URLString: url
            }
            document.getElementById("indexResults").innerHTML = "Indexing " + url + "..."
            document.getElementById("indexErrors").innerHTML = ""
            $.ajax({
                url: 'http://localhost:7250/api/crawl',
                type: 'post',
                dataType: 'json',
                contentType: 'application/json',
                success: function (data) {
                    var txt = "<table border='1'>"
                    txt += "<tr><th>Pages</th><th>Words Indexed</th><th>Duration (s)</th></tr>"
                    txt += "<tr><td>" + data.PagesCrawled + "</td><td>"+ data.WordsIndexed +"</td><td>"+ data.DurationSeconds+"</td></tr>"
                    txt += "</table>"
                    document.getElementById("indexResults").innerHTML = txt; 
                    if (data.CrawlErrors){
                        var errTxt = "<table border='1'><tr><th>Errors</th></tr>"
                        for (result of data.CrawlErrors) {
                            errTxt += "<tr><td>"+ result + "</td></tr>"
                        }
                        errTxt += "</table>"
                        document.getElementById("indexErrors").innerHTML = errTxt; 

                    }
                },
                error: function (err) {
                    document.getElementById("indexResults").innerHTML = "An error occurred indexing the page. Make sure your index url starts with http(s)://";                      

                },               
                data: JSON.stringify(search)
            });            
        });   
        $('#reset').click(function() {
            document.getElementById("resetResults").innerHTML = "Resetting the index..."
            $.ajax({
                url: 'http://localhost:7250/api/reset',
                type: 'DELETE',
                contentType:'application/json',  
                dataType: 'text',                
                success: function () {
                    document.getElementById("resetResults").innerHTML = "Index has been reset";                      
                },
                error: function () {
                    document.getElementById("resetResults").innerHTML = "An error occurred resetting the index";                      
                }
            });            
        });              
  });