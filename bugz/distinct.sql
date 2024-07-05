SELECT DISTINCT
    json_extract(OtherFieldsJSON, '$.assigned_to') AS Email,
    json_extract(OtherFieldsJSON, '$.assigned_to_detail.name') AS Name,
    json_extract(OtherFieldsJSON, '$.assigned_to_detail.real_name') AS RealName
FROM bugs
UNION
SELECT DISTINCT
    json_each.value AS Email,
    json_each.value AS Name,
    json_each.value AS RealName
FROM bugs, json_each(json(bugs.OtherFieldsJSON), '$.cc')
UNION
SELECT DISTINCT
    json_extract(OtherFieldsJSON, '$.cc_detail.email') AS Email,
    json_extract(OtherFieldsJSON, '$.cc.name') AS Name,
    json_extract(OtherFieldsJSON, '$.cc_detail.real_name') AS RealName
FROM bugs
UNION
SELECT DISTINCT
    json_extract(OtherFieldsJSON, '$.creator_detail.email') AS Email,
    json_extract(OtherFieldsJSON, '$.creator_detail.name') AS Name,
    json_extract(OtherFieldsJSON, '$.creator_detail.real_name') AS RealName
FROM bugs
UNION
SELECT DISTINCT
    json_extract(OtherFieldsJSON, '$.flags.requestee') AS Email,
    json_extract(OtherFieldsJSON, '$.flags.setter') AS Email,
    json_extract(OtherFieldsJSON, '$.flags.name') AS Name
FROM bugs
UNION
SELECT DISTINCT
    json_extract(OtherFieldsJSON, '$.qa_contact_detail.email') AS Email,
    json_extract(OtherFieldsJSON, '$.qa_contact_detail.name') AS Name,
    json_extract(OtherFieldsJSON, '$.qa_contact_detail.real_name') AS RealName
FROM bugs
UNION
SELECT DISTINCT
    Creator AS Email,
    Creator AS Name,
    Creator AS RealName
FROM comments
UNION
SELECT DISTINCT
    Creator AS Email,
    Creator AS Name,
    Creator AS RealName
FROM attachments;