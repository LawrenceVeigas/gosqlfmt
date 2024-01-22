## gosqlfmt

### To do

-

### Rough space

#### get tables

```
where
	a.dt between ${hiveconf:start_date} and ${hiveconf:end_date}
	and a.status = 'COMPLETED'
```

```
lateral view explode (a.discount_bifurcation :restaurant_funded_discount :offer_ids) d
left join excel.do_city_mapping b
on a.cityid = b.id
left join excel.do_dm_users c
on a.customer_id = c.user_id
```

- (?i)from(.\*)(?=where|group by|order by|having|qualify)
- (?i)from(.\*)
