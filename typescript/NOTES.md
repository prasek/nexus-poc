- Why not just HTTP and add a fetch() method in workflow?
  Definitely more practical...

```
// Start an ALO
curl -XPOST https://example.com/payments/af06ab3d
201 CREATED
{"status": "started"}
//
curl https://example.com/payments/af06ab3d
```

- TODO: Define which metrics we need to add
