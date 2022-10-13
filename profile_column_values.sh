#! /bin/bash
echo "Dropping the existing stats table (data_profiling.count_by_value) if present"
~/Documents/binary/clickhouse/clickhouse client --host expanse.localdomain --port 9002 --query 'DROP TABLE  IF EXISTS data_profiling.count_by_value'

echo "Creating an empty table for stats (data_profiling.count_by_value)"
~/Documents/binary/clickhouse/clickhouse client --host expanse.localdomain --port 9002 --query='CREATE TABLE data_profiling.count_by_value ( `table_name` String, `column_name` String, `data_type` String, `event_dataset` Nullable(String), `value` Nullable(String), `record_count` UInt64, `first_seen` Nullable(DateTime), `last_seen` Nullable(DateTime) ) ENGINE = MergeTree ORDER BY tuple() SETTINGS index_granularity = 8192;'

printf "\n\n**************************************** Counting Column By Value ****************************************\n\n"

while IFS="," read -r database_col table_col col_name data_type
do
  DB=$(eval echo $database_col) 
  TBL=$(eval echo $table_col) 
  COL=$(eval echo $col_name) 
  DTYPE=$(eval echo $data_type) 
  echo "Database: ${DB}    Table: ${TBL}    Data Type: ${DTYPE}     Column: ${COL}"
  ~/Documents/binary/clickhouse/clickhouse client --host expanse.localdomain --port 9002 --query="INSERT INTO data_profiling.count_by_value SELECT '${DB}.${TBL}' AS table_name, '${COL}' AS column_name, '${DTYPE}' AS data_type, v.event_dataset, v.${col_name} AS value, count(*) AS record_count, min(v.elk_ts) AS first_seen,  max(v.elk_ts) AS last_seen FROM public.onion_logs_v2 v GROUP BY v.event_dataset, v.${col_name} ORDER BY v.event_dataset, v.${col_name};"
  #break 
done < <(tail -n +2 v2_onion_log_schema_low_cardinality.csv)

printf "\n\n************************************************************************************************************************\n\n"
printf "Finished."
printf "\n\n************************************************************************************************************************\n\n"
