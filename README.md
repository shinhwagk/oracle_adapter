# oracle_adapter

- tables
```
 Schema |      Name      | Type  |  Owner   
--------+----------------+-------+----------
 public | metrics_copy   | table | postgres
 public | metrics_labels | table | postgres
 public | metrics_values | table | postgres
```

- table desc
```
 Column |    Type     | Collation | Nullable | Default 
--------+-------------+-----------+----------+---------
 sample | prom_sample |           | not null | 
Triggers:
    insert_trigger BEFORE INSERT ON metrics_copy FOR EACH ROW EXECUTE PROCEDURE prometheus.insert_view_normal('metrics_values', 'metrics_labels')
```

- table desc
```
   Column    |  Type   | Collation | Nullable |                  Default                   
-------------+---------+-----------+----------+--------------------------------------------
 id          | integer |           | not null | nextval('metrics_labels_id_seq'::regclass)
 metric_name | text    |           | not null | 
 labels      | jsonb   |           |          | 
Indexes:
    "metrics_labels_pkey" PRIMARY KEY, btree (id)
    "metrics_labels_metric_name_labels_key" UNIQUE CONSTRAINT, btree (metric_name, labels)
    "metrics_labels_labels_idx" gin (labels)
    "metrics_labels_metric_name_idx" btree (metric_name)
```

- table desc
```
  Column   |           Type           | Collation | Nullable | Default 
-----------+--------------------------+-----------+----------+---------
 time      | timestamp with time zone |           | not null | 
 value     | double precision         |           |          | 
 labels_id | integer                  |           |          | 
Indexes:
    "metrics_values_labels_id_idx" btree (labels_id, "time" DESC)
    "metrics_values_time_idx" btree ("time" DESC)
Number of child tables: 1 (Use \d+ to list them.)
```

- index
```
 Schema |                 Name                  | Type  |  Owner   |     Table      
--------+---------------------------------------+-------+----------+----------------
 public | metrics_labels_labels_idx             | index | postgres | metrics_labels
 public | metrics_labels_metric_name_idx        | index | postgres | metrics_labels
 public | metrics_labels_metric_name_labels_key | index | postgres | metrics_labels
 public | metrics_labels_pkey                   | index | postgres | metrics_labels
 public | metrics_values_labels_id_idx          | index | postgres | metrics_values
 public | metrics_values_time_idx               | index | postgres | metrics_values
```
