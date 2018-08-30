CREATE USER prometheus identified by prometheus;

DROP TABLE prometheus.tsdb_metrics_labels;
CREATE TABLE prometheus.tsdb_metrics_labels(
  id number DEFAULT prometheus.seq_tml.nextval not null,
  name varchar2(100) not null,
  labels varchar2(500)
)
PARTITION BY LIST (name) AUTOMATIC ( partition pc values ('pc') );

CREATE SEQUENCE prometheus.seq_tml cache 1000;
alter table prometheus.tsdb_metrics_labels compress;
alter table prometheus.tsdb_metrics_labels pctfree 0;
alter table prometheus.tsdb_metrics_labels pctused 99;
-- alter table prometheus.tsdb_metrics_labels ADD CONSTRAINT ensure_json CHECK(labels IS JSON);
-- alter table prometheus.tsdb_metrics_labels ADD CONSTRAINT unique_metric UNIQUE(name, labels) using index local;
create index prometheus.idx_tml_id on prometheus.tsdb_metrics_labels(id) local;
-- CREATE SEARCH INDEX prometheus.tml_labels_search_idx ON prometheus.tsdb_metrics_labels(labels) FOR JSON parameters('sync (EVERY 01:00:00)');

DROP TABLE prometheus.tsdb_metrics_values;
CREATE TABLE prometheus.tsdb_metrics_values(
  timestamp timestamp,
  value number,
  metric_id number
)
PARTITION BY RANGE (timestamp) INTERVAL ( NUMTODSINTERVAL (1, 'hour') )
(PARTITION pc VALUES LESS THAN (TO_DATE('1-7-1999', 'DD-MM-YYYY')));

-- ALTER TABLE prometheus.tsdb_metrics_values ADD CONSTRAINT fk_metric_id FOREIGN KEY(metric_id) REFERENCES prometheus.tsdb_metrics_labels(id);
CREATE INDEX prometheus.tmv_metric_id_idx ON prometheus.tsdb_metrics_values(metric_id) local;


create or replace PROCEDURE insert_metrics(nn  in varchar2,
                                           ll  in varchar2,
                                           ts  in timestamp,
                                           val in number) as
  metric_id number;
begin
  SELECT nvl((SELECT id
               FROM tsdb_metrics_labels
              WHERE name = nn
                and labels = ll),
             0)
    INTO metric_id
    FROM dual;
  if metric_id = 0 then
    select seq_tml.nextval into metric_id from dual;
    insert into tsdb_metrics_labels values (metric_id, nn, ll);
  end if;

  insert into tsdb_metrics_values values (ts, val, metric_id);

  COMMIT WRITE BATCH;
end;
/


create table test (name varchar2(100),labels varchar2(500),ts timestamp,val number);
