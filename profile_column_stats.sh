#! /bin/bash

echo "Getting the full list of columns"
~/Documents/binary/clickhouse/clickhouse client --host ${HOST_NAME} --port ${PORT} --query "SELECT t.database, t.table, t.name, t.type FROM system.columns t WHERE t.database = 'public' AND t.table = 'onion_logs_v2' ORDER BY t.type, t.name FORMAT CSVWithNames;" > v2_onion_log_schema_full.csv
echo "Dropping the existing stats table (data_profiling.column_statistics) if present"
~/Documents/binary/clickhouse/clickhouse client --host ${HOST_NAME} --port ${PORT} --query 'DROP TABLE  IF EXISTS data_profiling.column_statistics'

echo "Creating an empty table for stats (data_profiling.column_statistics)"
~/Documents/binary/clickhouse/clickhouse client --host ${HOST_NAME} --port ${PORT} --query='CREATE TABLE data_profiling.column_statistics ( `column_name` String, `event_dataset` Nullable(String), `count_distinct_values` UInt64, `count_not_null` UInt64, `count_of_values` UInt64, `count_of_null` Int64, `percent_null` Float64, `percent_not_null` Float64, `percent_unique` Nullable(Float64) ) ENGINE = MergeTree ORDER BY tuple() SETTINGS index_granularity = 8192;'

printf "\n\n**************************************** Calculating Column Stats ****************************************\n\n"
while IFS="," read -r database_col table_col col_name data_type
do
  DB=$(eval echo $database_col) 
  TBL=$(eval echo $table_col) 
  COL=$(eval echo $col_name) 
  DTYPE=$(eval echo $data_type) 
  echo "Database: ${DB}    Table: ${TBL}    Data Type: ${DTYPE}     Column: ${COL}"
  ~/Documents/binary/clickhouse/clickhouse client --host ${HOST_NAME} --port ${PORT} --query="INSERT INTO data_profiling.column_statistics SELECT '${COL}' AS column_name, event_dataset, COUNT(DISTINCT ${col_name}) AS count_distinct_values, COUNT(${col_name}) AS count_not_null, count(*) AS count_of_values, (count_of_values - count_not_null) AS count_of_null, (count_of_null   / count_of_values * 100.0 ) AS percent_null, (count_not_null  / count_of_values * 100.0 ) AS percent_not_null, CASE WHEN count_not_null = 0 THEN NULL ELSE (count_distinct_values / count_not_null  * 100.0 ) END AS percent_unique FROM public.onion_logs_v2 GROUP BY column_name, event_dataset ORDER BY column_name, event_dataset;"
  #break 
done < <(tail -n +2 v2_onion_log_schema_full.csv)

printf "\n\n************************************************************************************************************************\n\n"
printf "Finished."
printf "\n\n************************************************************************************************************************\n\n"
