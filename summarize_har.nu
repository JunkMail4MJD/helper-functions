print "   "
print "   "

let TODAY =  ^date -Iseconds | str substring 0..9
print "Today is: $TODAY"

let pp = ls *.har | sort-by type name | where type == 'file'

print $pp

print "   "
print "======================================================= Selecting file to reprocess ======================================================="
print "   "
let selected = $pp | input list --display name 'Select file to reprocess.'

print "SELECTED: "
print $selected
print "   "
print "   "

let outdb = $selected.name + '.db'

jq -Sc '.log.entries[] ' $selected.name
    | zq -j 'cut started_date_time:=startedDateTime, server_ip_address:=serverIPAddress, duration:=time, resource_type:=_resourceType, req_host:=parse_uri(request.url).host, req_path:=parse_uri(request.url).path, req_method:=request.method, req_http_ver:=request.httpVersion, req_body_size:=request.bodySize, res_mime_type:=response.content.mimeType, res_content_size:=response.content.size, resp_status:=response.status ' -
    | from json --objects
    | into sqlite $outdb --table-name input

duckdb -box -c "FROM input;" $outdb
