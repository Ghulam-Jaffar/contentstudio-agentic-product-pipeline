# Social Media API (v1.1) 26.03.2026

## facebook

---

### profile

---

#### user

---

##### auto-update task

---

###### Profile auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/update?load_feed_posts=true&max_posts=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_feed_posts | true | Enable or disable fetching of posts from the feed section of the profile. |
| max_posts | 50 | We do not recommend setting this parameter to values > 300 as it can cause the update process to fail.
If comments fetching is enabled, every post with comments is loaded in a separate task and contributes to total mentions usage. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| from_date | 2020-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).

This option is only available for the feed section.
This parameter is available only for Facebook users and pages and not applicable for groups.
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| load_reviews_posts | true | Enable or disable fetching of posts from the reviews section of the profile.
Applicable for type page and user+ (If the user has the field is_additional_profile_plus=true) |
| load_mentions_posts | true | Enable or disable fetching of posts from the mentions section of the profile.
Only Facebook pages or users with is_additional_profile_plus=True have this section. The parameter ignored for groups. |
| load_video_posts | true | Enable or disable fetching of posts from the videos section of the profile.

Only Facebook pages or users with is_additional_profile_plus=True have this section. The parameter ignored for groups. |
| load_reels_posts | true | Enable or disable fetching of posts from the reels section of the profile.

Only Facebook pages or users have this section. The parameter ignored for groups. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| comments_type | all | Facebook comments type of sorting |
| load_reactors | true | Enable or disable fetching of reactors for every post. |
| max_reactors | 50 | Set limit for the number of reactors that will be fetched for every post. |
| load_shares | true | Enable or disable fetching of shares for every post. |
| max_shares | 50 | Set limit for the number of shares that will be fetched for every post. |
| load_followers | true | Enable or disable fetching of followers of the profile. |
| max_followers | 50 | Set a limit for the number of followers that will be fetched for the profile. |
| load_following | true | Enable or disable fetching of following of the profile. |
| max_following | 50 | Set a limit for the number of following that will be fetched for the profile. |
| load_contact_info | true | Enable or disable fetching of contact info from the about section of the profile.

+1 credit per profile |
| load_page_transparency | true | Enable or disable fetching of page transparency info from the about section of the profile.

+3 mention per profile
Only Facebook pages or users with is_additional_profile_plus=True have this section. The parameter ignored for groups. |
| region | Indonesia%20-%20Indonesian | Set region for the profile to fetch localized data.

The available regions vary depending on the profile. To get localized data, set the region parameter and request the profile using the regional username or user ID.

You can find the list of available regions in the profile page under the Switch Region menu. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| upload_cover_photo_to_s3 | true | Upload profile cover photo to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot.

+15 mentions per profile. |
| upload_posts_screenshots_to_s3 | true | Take a screenshot for each item in the feed section of the profile.

+5 mentions per post. |
| upload_posts_images_to_s3 | true | Upload all posts images to external data storage.

+1 mention per image. |
| analyze_demography | true | Enable or disable fetching of the age and gender for the profile owner. |
| analyze_languages | true | Analyze profile languages

+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b1 | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve:
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items within a regular time-span. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/tasks/:task_id`

---

##### Profile update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/update?load_feed_posts=true&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_feed_posts | true | Enable or disable fetching of posts from the feed section of the profile. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the profile.

We do not recommend setting this parameter to values > 300 as it can cause the update process to fail.
If comments fetching is enabled, every post with comments is loaded in a separate task and contributes to total mentions usage. |
| from_date | 2020-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).

This option is only available for the feed section.
This parameter is available only for Facebook users and pages and not applicable for groups.
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| load_reviews_posts | true | Enable or disable fetching of posts from the reviews section of the profile.
Applicable for type page and user+ (If the user has the field is_additional_profile_plus=true) |
| load_mentions_posts | true | Enable or disable fetching of posts from the mentions section of the profile.

Only Facebook pages or users with is_additional_profile_plus=True have this section. The parameter ignored for groups. |
| load_video_posts | true | Enable or disable fetching of posts from the videos section of the profile.

Only Facebook pages or users with is_additional_profile_plus=True have this section. The parameter ignored for groups. |
| load_reels_posts | true | Enable or disable fetching of posts from the reels section of the profile.

Only Facebook pages or users have this section. The parameter ignored for groups. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| comments_type | all | Facebook comments type of sorting |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_reactors | true | Enable or disable fetching of reactors for every post. |
| max_reactors | 50 | Set limit for the number of reactors that will be fetched for every post. |
| load_shares | true | Enable or disable fetching of shares for every post. |
| max_shares | 50 | Set limit for the number of shares that will be fetched for every post. |
| load_followers | true | Enable or disable fetching of followers of the profile. |
| max_followers | 50 | Set a limit for the number of followers that will be fetched for the profile. |
| load_following | true | Enable or disable fetching of following of the profile. |
| max_following | 50 | Set a limit for the number of following that will be fetched for the profile. |
| load_contact_info | true | Enable or disable fetching of contact info from the about section of the profile.

+1 mention per profile |
| load_page_transparency | true | Enable or disable fetching of page transparency info from the about section of the profile.

+3 mention per profile
Only Facebook pages or users with is_additional_profile_plus=True have this section. The parameter ignored for groups. |
| region | Indonesia%20-%20Indonesian | Set region for the profile to fetch localized data.

The available regions vary depending on the profile. To get localized data, set the region parameter and request the profile using the regional username or user ID.

You can find the list of available regions in the profile page under the Switch Region menu. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| upload_cover_photo_to_s3 | true | Upload profile cover photo to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot (without loading media).
+5 mentions per profile. |
| upload_posts_screenshots_to_s3 | true | Take a screenshot for each item in the feed section of the profile.

+5 mentions per post. |
| upload_posts_images_to_s3 | true | Upload all posts images to external data storage.
+1 mention per image. |
| analyze_demography | true | Enable or disable fetching of the age and gender for the profile owner. |
| analyze_languages | true | Analyze profile languages

+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the profile task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/update`

---

##### Cached profile data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id`

---

##### Cached profile posts (feed)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/feed/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.
A query can contain quotes, logical operator "OR" and stopwords.
Query language can be chosen with the parameter "query_lang". This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cached profile posts (reviews)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/reviews/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.
A query can contain quotes, logical operator "OR" and stopwords.
Query language can be chosen with the parameter "query_lang". This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cached profile posts (mentions)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/mentions/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.
A query can contain quotes, logical operator "OR" and stopwords.
Query language can be chosen with the parameter "query_lang". This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cached profile posts (video)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/video/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.
A query can contain quotes, logical operator "OR" and stopwords.
Query language can be chosen with the parameter "query_lang". This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cached profile posts (reels)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/reels/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page. [1 .. 100] |
| from_date | 2021-01-01 | All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.
A query can contain quotes, logical operator "OR" and stopwords.
Query language can be chosen with the parameter "query_lang". This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cached profile friends (followers profiles)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/followers/profiles?max_page_size=50&order_by=id_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cached profile friends (following profiles)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/following/profiles?max_page_size=50&order_by=id_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### page

---

##### auto-update task

---

###### Profile auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/update?auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00&load_feed_posts=true&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_feed_posts | true | Enable or disable fetching of posts from the feed section of the profile. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the profile.
We do not recommend setting this parameter to values > 300 as it can cause the update process to fail.
If comments fetching is enabled, every post with comments is loaded in a separate task and contributes to total mentions usage. |
| from_date | 2020-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).

This option is only available for the feed section.
This parameter is available only for Facebook users and pages and not applicable for groups.
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| load_community_posts | true | Enable or disable fetching of posts from the community section of the profile. |
| load_reviews_posts | true | Enable or disable fetching of posts from the reviews section of the profile.
Applicable for type page and user+ (If the user has the field is_additional_profile_plus=true) |
| load_mentions_posts | true | Enable or disable fetching of posts from the mentions section of the profile.
Only Facebook pages or users with is_additional_profile_plus=True have this section. The parameter ignored for groups. |
| load_video_posts | true | Enable or disable fetching of posts from the videos section of the profile.

Only Facebook pages or users with is_additional_profile_plus=True have this section. The parameter ignored for groups. |
| load_reels_posts | true | Enable or disable fetching of posts from the reels section of the profile.

Only Facebook pages or users have this section. The parameter ignored for groups. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| comments_type | all | Facebook comments type of sorting |
| load_reactors | true | Enable or disable fetching of reactors for every post. |
| max_reactors | 50 | Set limit for the number of reactors that will be fetched for every post. |
| load_shares | true | Enable or disable fetching of shares for every post. |
| max_shares | 50 | Set limit for the number of shares that will be fetched for every post. |
| load_contact_info | true | Enable or disable fetching of contact info from the about section of the profile.

+1 mention per profile |
| load_page_transparency | true | Enable or disable fetching of page transparency info from the about section of the profile.

+3 credits per profile
Only Facebook pages or users with is_additional_profile_plus=True have this section. The parameter ignored for groups. |
| region | Indonesia%20-%20Indonesian | Set region for the profile to fetch localized data.

The available regions vary depending on the profile. To get localized data, set the region parameter and request the profile using the regional username or user ID.

You can find the list of available regions in the profile page under the Switch Region menu. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| upload_cover_photo_to_s3 | true | Upload profile cover photo to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot (without loading media).
+5 mentions per profile. |
| upload_posts_screenshots_to_s3 | true | Take a screenshot for each item in the feed section of the profile.

+5 mentions per post. |
| upload_posts_images_to_s3 | true | Upload all posts images to external data storage.

+1 mention per image. |
| analyze_demography | true | Enable or disable fetching of the age and gender for the profile owner. |
| analyze_languages | true | Analyze profile languages

+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b1 | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve:
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items within a regular time-span. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/tasks/:task_id`

---

##### Profile update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/update?load_feed_posts=true&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_feed_posts | true | Enable or disable fetching of posts from the feed section of the profile. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the profile.
We do not recommend setting this parameter to values > 300 as it can cause the update process to fail.
If comments fetching is enabled, every post with comments is loaded in a separate task and contributes to total mentions usage. |
| from_date | 2020-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).

This option is only available for the feed section.
This parameter is available only for Facebook users and pages and not applicable for groups.
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| load_community_posts | true | Enable or disable fetching of posts from the community section of the profile. |
| load_reviews_posts | true | Enable or disable fetching of posts from the reviews section of the profile.
Applicable for type page and user+ (If the user has the field is_additional_profile_plus=true) |
| load_mentions_posts | true | Enable or disable fetching of posts from the mentions section of the profile.

Only Facebook pages or users with is_additional_profile_plus=True have this section. The parameter ignored for groups. |
| load_video_posts | true | Enable or disable fetching of posts from the videos section of the profile.
Only Facebook pages or users with is_additional_profile_plus=True have this section. The parameter ignored for groups. |
| load_reels_posts | true | Enable or disable fetching of posts from the reels section of the profile.

Only Facebook pages or users have this section. The parameter ignored for groups. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| comments_type | all | Facebook comments type of sorting |
| load_reactors | true | Enable or disable fetching of reactors for every post. |
| max_reactors | 50 | Set limit for the number of reactors that will be fetched for every post. |
| load_shares | true | Enable or disable fetching of shares for every post. |
| max_shares | 50 | Set limit for the number of shares that will be fetched for every post. |
| load_contact_info | true | Enable or disable fetching of contact info from the about section of the profile.

+1 mention per profile |
| load_page_transparency | true | Enable or disable fetching of page transparency info from the about section of the profile.

+3 mention per profile
Only Facebook pages or users with is_additional_profile_plus=True have this section. The parameter ignored for groups. |
| region | Indonesia%20-%20Indonesian | Set region for the profile to fetch localized data.

The available regions vary depending on the profile. To get localized data, set the region parameter and request the profile using the regional username or user ID.

You can find the list of available regions in the profile page under the Switch Region menu. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| upload_cover_photo_to_s3 | true | Upload profile cover photo to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot (without loading media).
+5 mentions per profile. |
| upload_posts_screenshots_to_s3 | true | Take a screenshot for each item in the feed section of the profile.

+5 mentions per post. |
| upload_posts_images_to_s3 | true | Upload all posts images to external data storage.

+1 mention per image. |
| analyze_demography | true | Enable or disable fetching of the age and gender for the profile owner. |
| analyze_languages | true | Analyze profile languages

+1 mention per profile |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the profile task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/update`

---

##### Cached profile data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id`

---

##### Cached profile posts (feed)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/feed/posts?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cached profile posts (reviews)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/reviews/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.
A query can contain quotes, logical operator "OR" and stopwords.
Query language can be chosen with the parameter "query_lang". This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cached profile posts (community)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/community/posts?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cached profile posts (mentions)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/mentions/posts?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cached profile posts (video)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/video/posts?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cached profile posts (reels)

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/reels/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page. [1 .. 100] |
| from_date | 2021-01-01 | All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.
A query can contain quotes, logical operator "OR" and stopwords.
Query language can be chosen with the parameter "query_lang". This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### group

---

##### auto-update task

---

###### Profile auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/update?auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00&load_feed_posts=true&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_feed_posts | true | Enable or disable fetching of posts from the feed section of the profile. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the profile.
We do not recommend setting this parameter to values > 300 as it can cause the update process to fail.
If comments fetching is enabled, every post with comments is loaded in a separate task and contributes to total mentions usage. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| comments_type | all | Facebook comments type of sorting |
| load_reactors | true | Enable or disable fetching of reactors for every post. |
| max_reactors | 50 | Set limit for the number of reactors that will be fetched for every post. |
| load_shares | true | Enable or disable fetching of shares for every post. |
| max_shares | 50 | Set limit for the number of shares that will be fetched for every post. |
| load_contact_info | true | Enable or disable fetching of contact info from the about section of the profile.

+1 mention per profile |
| region | Indonesia%20-%20Indonesian | Set region for the profile to fetch localized data.

The available regions vary depending on the profile. To get localized data, set the region parameter and request the profile using the regional username or user ID.

You can find the list of available regions in the profile page under the Switch Region menu. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| upload_cover_photo_to_s3 | true | Upload profile cover photo to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot (without loading media).
+5 mentions per profile. |
| upload_posts_screenshots_to_s3 | true | Take a screenshot for each item in the feed section of the profile.

+5 mentions per post. |
| upload_posts_images_to_s3 | true | Upload all posts images to external data storage.

+1 mention per image. |
| analyze_demography | true | Enable or disable fetching of the age and gender for the profile owner. |
| analyze_languages | true | Analyze profile languages

+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve:
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items within a regular time-span. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/tasks/:task_id`

---

##### Profile update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/update?load_feed_posts=true&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_feed_posts | true | Enable or disable fetching of posts from the feed section of the profile. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the profile.
We do not recommend setting this parameter to values > 300 as it can cause the update process to fail.
If comments fetching is enabled, every post with comments is loaded in a separate task and contributes to total mentions usage. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| comments_type | all | Facebook comments type of sorting |
| load_reactors | true | Enable or disable fetching of reactors for every post. |
| max_reactors | 50 | Set limit for the number of reactors that will be fetched for every post. |
| load_shares | true | Enable or disable fetching of shares for every post. |
| max_shares | 50 | Set limit for the number of shares that will be fetched for every post. |
| load_contact_info | true | Enable or disable fetching of contact info from the about section of the profile.

+1 mention per profile |
| region | Indonesia%20-%20Indonesian | Set region for the profile to fetch localized data.

The available regions vary depending on the profile. To get localized data, set the region parameter and request the profile using the regional username or user ID.

You can find the list of available regions in the profile page under the Switch Region menu. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| upload_cover_photo_to_s3 | true | Upload profile cover photo to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot.

+15 mentions per profile. |
| upload_posts_screenshots_to_s3 | true | Take a screenshot for each item in the feed section of the profile.

+5 mentions per post. |
| upload_posts_images_to_s3 | true | Upload all posts images to external data storage.

+1 mention per image. |
| analyze_demography | true | Enable or disable fetching of the age and gender for the profile owner. |
| analyze_languages | true | Analyze profile languages

+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the profile task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/update`

---

##### Cached profile data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id`

---

##### Cached profile posts

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/profile/:profile_id/feed/posts?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### posts search

---

#### for latest posts by keyword

---

##### auto-update task

---

###### Search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/posts/latest/update?auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| max_posts | 50 | Set a limit for the number of posts that will be fetched for the search.
Every post loaded in a separate task and contributes to total mentions usage. |
| location_id | 106078429431815 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| author_id | 10432356007 | Filter items by the author.
The parameter must be set to an internal Facebook ID of the target profile. |
| from_date | 2020-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| comments_type | all | Facebook comments type of sorting |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_reactors | true | Enable or disable fetching of reactors for every post. |
| max_reactors | 50 | Set a limit for the number of reactors that will be fetched for every post. |
| load_shares | true | Enable or disable fetching of shares for every post. |
| max_shares | 50 | Set a limit for the number of shares that will be fetched for every post. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/posts/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve:
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items within a regular time-span. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/posts/tasks/:task_id`

---

##### Task to update posts search

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/posts/latest/update?max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_posts | 50 | Set a limit for the number of posts that will be fetched for the search.
Every post loaded in a separate task and contributes to total mentions usage. |
| location_id | 106078429431815 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| author_id | 10432356007 | Filter items by the author.
The parameter must be set to an internal Facebook ID of the target profile. |
| from_date | 2021-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).

Timestamp value must be in ISO 8601 format.
All items created after and at this date will be removed from the response. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| comments_type | all | Facebook comments type of sorting |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_reactors | true | Enable or disable fetching of reactors for every post. |
| max_reactors | 50 | Set a limit for the number of reactors that will be fetched for every post. |
| load_shares | true | Enable or disable fetching of shares for every post. |
| max_shares | 50 | Set a limit for the number of shares that will be fetched for every post. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/posts/latest/update`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| location_id | 106078429431815 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| author_id | 10432356007 | Filter items by the author.
The parameter must be set to an internal Facebook ID of the target profile. |
| from_date | 2020-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |

---

##### Cached posts from search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/posts/latest/posts?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| location_id | 106078429431815 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| author_id | 10432356007 | Filter items by the author.
The parameter must be set to an internal Facebook ID of the target profile. |
| from_date | 2020-01-01 | All items created before this date will be removed from the response. |
| to_date | 2020-05-01 | All items created after and at this date will be removed from the response. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.

A query can contain quotes, logical operator OR and stopwords. |
| owner_types | user,page,group | Filter posts by the owner"s profile type.
Use "," to separate multiple values. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### for top posts by keyword

---

##### auto-update task

---

###### Search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/posts/top/update?auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| max_posts | 50 | Set a limit for the number of posts that will be fetched for the search.
Every post loaded in a separate task and contributes to total mentions usage. |
| location_id | 106078429431815 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| author_id | 10432356007 | Filter items by the author.
The parameter must be set to an internal Facebook ID of the target profile. |
| from_date | 2020-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| comments_type | all | Facebook comments type of sorting |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_reactors | true | Enable or disable fetching of reactors for every post. |
| max_reactors | 50 | Set a limit for the number of reactors that will be fetched for every post. |
| load_shares | true | Enable or disable fetching of shares for every post. |
| max_shares | 50 | Set a limit for the number of shares that will be fetched for every post. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/posts/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve:
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items within a regular time-span. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/posts/tasks/:task_id`

---

##### Task to update posts search

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/posts/top/update?max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_posts | 50 | Set a limit for the number of posts that will be fetched for the search.
Every post loaded in a separate task and contributes to total mentions usage. |
| location_id | 106078429431815 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| author_id | 10432356007 | Filter items by the author.
The parameter must be set to an internal Facebook ID of the target profile. |
| from_date | 2021-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| comments_type | all | Facebook comments type of sorting |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_reactors | true | Enable or disable fetching of reactors for every post. |
| max_reactors | 50 | Set a limit for the number of reactors that will be fetched for every post. |
| load_shares | true | Enable or disable fetching of shares for every post. |
| max_shares | 50 | Set a limit for the number of shares that will be fetched for every post. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/posts/top/update`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| location_id | 106078429431815 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| author_id | 10432356007 | Filter items by the author.
The parameter must be set to an internal Facebook ID of the target profile. |
| from_date | 2020-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |

---

##### Cached posts from search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/posts/top/posts?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| location_id | 106078429431815 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| author_id | 10432356007 | Filter items by the author.
The parameter must be set to an internal Facebook ID of the target profile. |
| from_date | 2020-01-01 | All items created before this date will be removed from the response. |
| to_date | 2020-05-01 | All items created after and at this date will be removed from the response. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.

A query can contain quotes, logical operator OR and stopwords. |
| owner_types | user,page,group | Filter posts by the owner"s profile type.
Use "," to separate multiple values. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### for posts by hashtag

---

##### auto-update task

---

###### Search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/posts/hashtag/update?max_posts=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_posts | 50 | Set a limit for the number of posts that will be fetched for the search.
Every post loaded in a separate task and contributes to total mentions usage. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| comments_type | all | Facebook comments type of sorting |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_reactors | true | Enable or disable fetching of reactors for every post. |
| max_reactors | 50 | Set a limit for the number of reactors that will be fetched for every post. |
| load_shares | true | Enable or disable fetching of shares for every post. |
| max_shares | 50 | Set a limit for the number of shares that will be fetched for every post. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/posts/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/posts/tasks/:task_id`

---

##### Task to update posts search

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/posts/hashtag/update?max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_posts | 50 | Set a limit for the number of posts that will be fetched for the search.
Every post loaded in a separate task and contributes to total mentions usage. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| comments_type | all | Facebook comments type of sorting |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_reactors | true | Enable or disable fetching of reactors for every post. |
| max_reactors | 50 | Set a limit for the number of reactors that will be fetched for every post. |
| load_shares | true | Enable or disable fetching of shares for every post. |
| max_shares | 50 | Set a limit for the number of shares that will be fetched for every post. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/posts/hashtag/update`

---

##### Cached posts from search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/posts/hashtag/posts?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2020-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| query | toyota -hybrid | Filter items by the text search query.

A query can contain quotes, logical operator OR and stopwords. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |
| owner_types | user,page,group | Filter posts by the owner"s profile type.
Use "," to separate multiple values. |

---

### ad posts search

---

#### ad posts search

---

##### auto-update task

---

###### Ad Posts search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/ad_posts/update?keywords=election&max_posts=50&auto_update_interval=300&auto_update_expire_at=2027-01-01T15:00:00&active_status=all&ad_type=all`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | election | Keywords to search for in the ad posts. |
| max_posts | 50 | Set limit for the number of ad posts that will be fetched for the search. |
| auto_update_interval | 300 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| active_status | all | Filter by the active status of the ad posts. |
| ad_type | all | Filter by ad type (e.g., political or issue ads). |
| content_languages | en,fr | Filter by content languages (ISO 639-1 codes). |
| countries | US,FR | Filter by countries (ISO 3166-1 alpha-2 codes). |
| media_type | all | Filter by media type of the ad creative. |
| page_ids | 1535230416709539,76283417065 | Filter by Facebook Page IDs associated with the ad posts. |
| publisher_platforms | facebook,instagram | Filter by publisher platforms where the ads were displayed. |
| min_start_date | 2024-01-01 | Minimum start date for the ad posts (ISO 8601 format YYYY-MM-DD). |
| max_start_date | 2024-05-01 | Maximum start date for the ad posts (ISO 8601 format YYYY-MM-DD). |
| sort_by_total_impressions | desc | Sort by total impressions of the ads. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/ad_posts/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

The parameter does not affect the number of items fetched for profiles, searches, etc during the updating process.
[ 1 .. 100 ] |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/ad_posts/tasks/:task_id`

---

##### Ad Posts search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/ad_posts/update?keywords=election&max_posts=50&active_status=all&ad_type=all`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | election | Keywords to search for in the ad posts. |
| max_posts | 50 | Set limit for the number of ad posts that will be fetched for the search. |
| active_status | all | Filter by the active status of the ad posts. |
| ad_type | all | Filter by ad type (e.g., political or issue ads). |
| content_languages | en,fr | Filter by content languages (ISO 639-1 codes). |
| countries | US,FR | Filter by countries (ISO 3166-1 alpha-2 codes). |
| media_type | all | Filter by media type of the ad creative. |
| page_ids | 1535230416709539,76283417065 | Filter by Facebook Page IDs associated with the ad posts. |
| publisher_platforms | facebook,instagram | Filter by publisher platforms where the ads were displayed. |
| min_start_date | 2024-01-01 | Minimum start date for the ad posts (ISO 8601 format YYYY-MM-DD). |
| max_start_date | 2024-05-01 | Maximum start date for the ad posts (ISO 8601 format YYYY-MM-DD) |
| sort_by_total_impressions | desc | Sort by total impressions of the ads. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the ad post search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/ad_posts/update?keywords=election&active_status=all&ad_type=all`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | election | Keywords to search for in the ad posts. |
| active_status | all | Filter by the active status of the ad posts. |
| ad_type | all | Filter by ad type (e.g., political or issue ads). |
| content_languages | en,fr | Filter by content languages (ISO 639-1 codes). |
| countries | US,FR | Filter by countries (ISO 3166-1 alpha-2 codes). |
| media_type | all | Filter by media type of the ad creative. |
| page_ids | 1535230416709539,76283417065 | Filter by Facebook Page IDs associated with the ad posts |
| publisher_platforms | facebook,instagram | Filter by publisher platforms where the ads were displayed. |
| min_start_date | 2024-01-01 | Minimum start date for the ad posts (ISO 8601 format YYYY-MM-DD). |
| max_start_date | 2024-05-01 | Maximum start date for the ad posts (ISO 8601 format YYYY-MM-DD) |
| sort_by_total_impressions | desc | Sort by total impressions of the ads. |

---

##### Cached posts from ad posts search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/ad_posts/items?keywords=election&max_page_size=50&order_by=id_desc&active_status=all&ad_type=all`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | election | Keywords to search for in the ad posts. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| active_status | all | Filter by the active status of the ad posts. |
| ad_type | all | Filter by ad type (e.g., political or issue ads). |
| content_languages | en,fr | Filter by content languages (ISO 639-1 codes). |
| countries | US,FR | Filter by countries (ISO 3166-1 alpha-2 codes). |
| media_type | all | Filter by media type of the ad creative. |
| page_ids | 1535230416709539,76283417065 | Filter by Facebook Page IDs associated with the ad posts. |
| publisher_platforms | facebook,instagram | Filter by publisher platforms where the ads were displayed. |
| min_start_date | 2024-01-01 | Minimum start date for the ad posts (ISO 8601 format YYYY-MM-DD). |
| max_start_date | 2024-05-01 | Maximum start date for the ad posts (ISO 8601 format YYYY-MM-DD). |
| sort_by_total_impressions | desc | Sort by total impressions of the ads. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### profiles search

---

#### for people

---

##### auto-update task

---

###### Search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/profiles/people/update?max_profiles=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_profiles | 50 | Set limit for the number of profiles that will be fetched for the search. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| location_id | 104597642912469 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| school_id | 102327309236334 | Filter items by the school.
The parameter must be set to an internal Facebook ID of the target school. |
| employer_id | 100484820802 | Filter items by the employer.
The parameter must be set to an internal Facebook ID of the target employer. |
| load_profiles_data | true | If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| upload_cover_photo_to_s3 | true | Upload profile cover photo to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot (without loading media).
+5 mentions per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/profiles/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve:
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items within a regular time-span. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/profiles/tasks/:task_id`

---

##### Task to update profiles search

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/profiles/people/update?max_profiles=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_profiles | 50 | Set limit for the number of profiles that will be fetched for the search. |
| location_id | 104597642912469 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| school_id | 102327309236334 | Filter items by the school.
The parameter must be set to an internal Facebook ID of the target school. |
| employer_id | 100484820802 | Filter items by the employer.
The parameter must be set to an internal Facebook ID of the target employer. |
| load_profiles_data | true | If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| upload_cover_photo_to_s3 | true | Upload profile cover photo to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot.
+15 mentions per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/profiles/people/update`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| location_id | 104597642912469 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| school_id | 102327309236334 | Filter items by the school.
The parameter must be set to an internal Facebook ID of the target school. |
| employer_id | 100484820802 | Filter items by the employer.
The parameter must be set to an internal Facebook ID of the target employer. |

---

##### Cached profiles from search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/profiles/people/profiles?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| location_id | 104597642912469 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |
| school_id | 102327309236334 | Filter items by the school.
The parameter must be set to an internal Facebook ID of the target school. |
| employer_id | 100484820802 | Filter items by the employer.
The parameter must be set to an internal Facebook ID of the target employer. |

---

#### for pages

---

##### auto-update task

---

###### Search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/profiles/pages/update?auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00&max_profiles=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| max_profiles | 50 | Set limit for the number of profiles that will be fetched for the search. |
| location_id | 104597642912469 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| category_id | 1006 | Filter items by the category. List of possible category_id values:
1006 (Local Business or Place)
1007 (Artist, Band or Public Figure)
1009 (Brand or Product)
1013 (Company, Organisation or Institution)
1019 (Entertainment)
2612 (Cause or Community) |
| is_shop | true | Filter items by the shop attribute. |
| load_profiles_data | true | If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| upload_cover_photo_to_s3 | true | Upload profile cover photo to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot (without loading media).
+5 mentions per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/profiles/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve:
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items within a regular time-span. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/profiles/tasks/:task_id`

---

##### Task to update profiles search

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/profiles/pages/update?max_profiles=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_profiles | 50 | Set limit for the number of profiles that will be fetched for the search. |
| location_id | 104597642912469 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| category_id | 1006 | Filter items by the category. List of possible category_id values:
1006 (Local Business or Place)
1007 (Artist, Band or Public Figure)
1009 (Brand or Product)
1013 (Company, Organisation or Institution)
1019 (Entertainment)
2612 (Cause or Community) |
| is_shop | true | Filter items by the shop attribute. |
| load_profiles_data | true | If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| upload_cover_photo_to_s3 | true | Upload profile cover photo to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot.
+15 mentions per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/profiles/pages/update`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| location_id | 104597642912469 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| category_id | 1006 | Filter items by the category. List of possible category_id values:
1006 (Local Business or Place)
1007 (Artist, Band or Public Figure)
1009 (Brand or Product)
1013 (Company, Organisation or Institution)
1019 (Entertainment)
2612 (Cause or Community) |
| is_shop | 1 | Filter items by the shop attribute. |

---

##### Cached profiles from search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/profiles/pages/profiles?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| location_id | 104597642912469 | Filter items by the tagged location.
The parameter must be set to an internal Facebook ID of the target location. |
| category_id | 1006 | Filter items by the category. List of possible category_id values:
1006 (Local Business or Place)
1007 (Artist, Band or Public Figure)
1009 (Brand or Product)
1013 (Company, Organisation or Institution)
1019 (Entertainment)
2612 (Cause or Community) |
| is_shop | 1 | Filter items by the shop attribute. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### for groups

---

##### auto-update task

---

###### Search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/profiles/groups/update?auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00&max_profiles=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| max_profiles | 50 | Set limit for the number of profiles that will be fetched for the search. |
| load_profiles_data | true | If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| upload_cover_photo_to_s3 | true | Upload profile cover photo to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot (without loading media).
+5 mentions per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/profiles/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve:
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items within a regular time-span. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/profiles/tasks/:task_id`

---

##### Task to update profiles search

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/profiles/groups/update?max_profiles=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_profiles | 50 | Set limit for the number of profiles that will be fetched for the search. |
| load_profiles_data | true | If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| upload_cover_photo_to_s3 | true | Upload profile cover photo to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot.
+15 mentions per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/profiles/groups/update`

---

##### Cached profiles from search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/search/:search_request/profiles/groups/profiles?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### post

---

#### auto-update task

---

##### Post auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/post/:post_id/update?auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00&load_comments=true&max_comments=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.
We do not recommend setting this parameter to values > 100 as it can cause the update process to fail.
Values < 25 will drastically increase the update process performance. |
| comments_type | all | Facebook comments type of sorting |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_reactors | true | Enable or disable fetching of reactors for the post. |
| max_reactors | 50 | Set limit for the number of reactors that will be fetched for the post. |
| load_shares | true | Enable or disable fetching of shares for the post. |
| max_shares | 50 | Set limit for the number of shares that will be fetched for the post. |
| load_mediaset | true | Enable or disable fetching the complete media set for posts with >5 images. +3 credits per post with >5 images. |
| upload_posts_screenshots_to_s3 | true | Take a screenshot of the post.

+5 mentions per post. |
| upload_posts_images_to_s3 | true | Upload all posts images to external data storage.

+1 mention per image. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve:
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items within a regular time-span. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/facebook/post/tasks/:task_id`

---

#### Post update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/facebook/post/:post_id/update?load_comments=true&max_comments=50&comments_type=all`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.
We do not recommend setting this parameter to values > 100 as it can cause the update process to fail.
Values < 25 will drastically increase the update process performance. |
| comments_type | all | Facebook comments type of sorting |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_reactors | true | Enable or disable fetching of reactors for the post. |
| max_reactors | 50 | Set limit for the number of reactors that will be fetched for the post. |
| load_shares | true | Enable or disable fetching of shares for the post. |
| max_shares | 50 | Set limit for the number of shares that will be fetched for the post. |
| load_mediaset | true | Enable or disable fetching the complete media set for posts with >5 images. +3 credits per post with >5 images. |
| upload_posts_screenshots_to_s3 | true | Take a screenshot of the post.

+5 mentions per post. |
| upload_posts_images_to_s3 | true | Upload all posts images to external data storage.

+1 mention per image. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process is finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the post task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/post/:post_id/update`

---

#### Cached post data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/post/:post_id`

---

#### Cached post comments

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/post/:post_id/comments?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached post reactors

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/post/:post_id/reactors/:reaction_type/profiles?max_page_size=50&order_by=id_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| cursor | aWRfZGVzY3wxMDAwMDAwMzM2NDAwNjA= | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached post shares

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/post/:post_id/shares/posts?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| query | toyota | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached comment replies

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/comment/:comment_id/replies?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### comment

---

#### Cached post comments

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/post/:post_id/comments?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached comment data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/comment/:comment_id`

---

#### Cached comment replies

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/comment/:comment_id/replies?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### Facebook queue size

**GET** `{{API_BASE_URL}}{{API_VERSION}}/facebook/queues/size`

---

### API usage stats

**GET** `{{API_BASE_URL}}{{API_VERSION}}/stats/usage?from_date=2025-03-25&to_date=2025-04-01&services=facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| from_date | 2025-03-25 | Calculate usage from this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| to_date | 2025-04-01 | Calculate usage to this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| services | facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest | Filter usage stats by services.

Multiple values can be provided as a comma-separated list. |

---

## instagram

---

### profile

---

#### auto-update task

---

##### Profile auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/:profile_id/update?load_feed_posts=true&max_posts=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_feed_posts | true | Fetch posts from /feed of a profile |
| max_posts | 50 | Maximum number of posts to fetch |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| from_date | 2021-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| load_reels_posts | true | Fetch posts from /reels page of a profile |
| load_tagged_posts | true | Fetch posts from /tagged page of a profile |
| load_stories | true | Fetch the latest story of a profile |
| load_highlights | true | Fetch highlight stories of a profile |
| max_stories | 50 | Maximum number of stories to fetch |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_followers | true | Enable or disable fetching of followers of the profile. |
| load_followers_data | true | Fetch all fields of followers profiles.

This parameter can be enabled only if load_followers parameter is enabled.
+9 mentions per follower. |
| max_followers | 50 | Set a limit for the number of followers that will be fetched for the profile. |
| load_following | true | Enable or disable fetching of following of the profile. |
| max_following | 50 | Set a limit for the number of following that will be fetched for the profile. |
| load_suggested | true | Enable or disable fetching of suggested profiles. |
| load_contact_data | true | Enable or disable fetching of contact data for the profile. |
| load_accessibility_caption | true | Enable or disable fetching of accessibility caption for the tag posts. Accessibility caption will be available in attached_media_content field. |
| load_usertags | true | Enable or disable fetching of usertags for the post. Tagged users will be available in attached_media_tagged_users field. |
| load_posts_data | true | Fetch all fields for tagged posts (including location_id, product_type and attached_carousel_media_urls). |
| load_about_account | true | Load data from 'About this account' menu.

+3 mentions per task. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| page_screenshot | true | Take full page screenshot.

+15 mentions per profile. |
| analyze_demography | true | Analyze age and gender for profile |
| analyze_languages | true | Analyze profile languages

+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/tasks/:task_id`

---

#### Profile update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/:profile_id/update?load_feed_posts=true&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_feed_posts | true | Fetch posts from /feed of a profile |
| max_posts | 50 | Maximum number of posts to fetch |
| from_date | 2021-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| load_reels_posts | true | Fetch posts from /reels page of a profile |
| load_tagged_posts | true | Fetch posts from /tagged page of a profile |
| load_stories | true | Fetch the latest story of a profile |
| load_highlights | true | Fetch highlight stories of a profile |
| max_stories | 50 | Maximum number of stories to fetch |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_followers | true | Enable or disable fetching of followers of the profile. |
| load_followers_data | true | Fetch all fields of followers profiles.

This parameter can be enabled only if load_followers parameter is enabled.
+9 mentions per follower. |
| max_followers | 50 | Set a limit for the number of followers that will be fetched for the profile. |
| load_following | true | Enable or disable fetching of following of the profile. |
| max_following | 50 | Set a limit for the number of following that will be fetched for the profile. |
| load_suggested | true | Enable or disable fetching of suggested profiles. |
| load_contact_data | true | Enable or disable fetching of contact data for the profile. |
| load_accessibility_caption | true | Enable or disable fetching of accessibility caption for the tag posts. Accessibility caption will be available in attached_media_content field. |
| load_usertags | true | Enable or disable fetching of usertags for the post. Tagged users will be available in attached_media_tagged_users field. |
| load_posts_data | true | Fetch all fields for tagged posts (including location_id, product_type and attached_carousel_media_urls). |
| load_about_account | true | Load data from 'About this account' menu.
+3 mentions per task. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot.

+15 mentions per profile. |
| analyze_demography | true | Analyze age and gender for profile |
| analyze_languages | true | Analyze profile languages

+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the profile task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/:profile_id/update`

---

#### Cached profile data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/:profile_id`

---

#### Cached profile feed posts

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/:profile_id/:section/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| location_name | New York | Filter items by the location name.
A query can contain quotes, logical operators, parenthesis and stopwords. |
| query | love | Filter items by the text search query.
A query can contain quotes, logical operator "OR" and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached profile reels posts

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/:profile_id/:section/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| location_name | New York | Filter items by the location name.
A query can contain quotes, logical operators, parenthesis and stopwords. |
| query | drawing | Filter items by the text search query.
A query can contain quotes, logical operator "OR" and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached profile tagged posts

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/:profile_id/:section/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| location_name | New York | Filter items by the location name.
A query can contain quotes, logical operators, parenthesis and stopwords. |
| query | drawing | Filter items by the text search query.
A query can contain quotes, logical operator "OR" and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached profile stories

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/:profile_id/:section/items?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MjY3NDI2NjQ0ODMxNzk4 | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached profile highlights

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/:profile_id/:section/items?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MjY3NDI2NjQ0ODMxNzk4 | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached followers profiles

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/:profile_id/followers/profiles?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached following profiles

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/:profile_id/following/profiles?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached suggested profiles

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/profile/:profile_id/suggested/profiles?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### profiles search

---

#### auto-update task

---

##### Search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/profiles/update?keywords=bill gates&max_profiles=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | bill gates | List of keywords or a phrase to search |
| max_profiles | 50 | Maximum number of profiles to fetch |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_profiles_data | true | If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/profiles/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/profiles/tasks/:task_id`

---

#### Task to update profiles search

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/profiles/update?keywords=bill gates&max_profiles=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | bill gates | List of keywords or a phrase to search |
| max_profiles | 50 | Maximum number of profiles to fetch |
| load_profiles_data | true | If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/profiles/update?keywords=bill gates`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | bill gates | List of keywords or a phrase to search |

---

#### Cached profiles from search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/profiles/items?keywords=bill gates&order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | bill gates | List of keywords or a phrase to search |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### posts search

---

#### for posts by hashtag

---

##### auto-update task

---

###### Search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/tag/:tag_id/update?load_posts=true&max_posts=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_posts | true | Fetch tagged posts |
| max_posts | 50 | Maximum number of posts to fetch |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| from_date | 2021-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| sort_type | recent | Sort type for tagged posts

"recent" - open recent posts (without reels)
"top" - open top posts (including reels)
"clips" - open video posts (only reels) |
| load_posts_data | true | Fetch all fields for tagged posts (including location_id, product_type and attached_carousel_media_urls).
Each post update will cost 2 mentions instead of 1 if you enable this option. |
| load_carousel_posts | true | Enable or disable fetching of carousel images for posts.
Every post that has carousel images will cost 2 to update instead of 1 if this option is enabled. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_accessibility_caption | true | Enable or disable fetching of accessibility caption for the tag posts. Accessibility caption will be available in attached_media_content field.

+1 mentions per tag post. |
| load_usertags | true | Enable or disable fetching of accessibility caption for the tag posts. Accessibility caption will be available in attached_media_content field.

+1 mentions per tag post. |
| upload_posts_image_to_s3 | true | Upload attached_media_display_url image to external data storage.

+1 credit per image. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/tag/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/instagram/tag/tasks/:task_id`

---

##### Task to update hashtag posts search

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/tag/:tag_id/update?load_posts=true&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_posts | true | Fetch tagged posts |
| max_posts | 50 | Maximum number of posts to fetch |
| from_date | 2021-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| sort_type | recent | Sort type for tagged posts

"recent" - open recent posts (without reels)
"top" - open top posts (including reels)
"clips" - open video posts (only reels) |
| load_posts_data | true | Fetch all fields for tagged posts (including owner_username, location_id and attached_carousel_media_urls).

If this option is enabled: every post update will cost 2 mentions instead of 1. |
| load_carousel_posts | true | Enable or disable fetching of carousel images for posts.
Every post that has carousel images will cost 2 to update instead of 1 if this option is enabled. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_accessibility_caption | true | Enable or disable fetching of accessibility caption for the tag posts. Accessibility caption will be available in attached_media_content field.

+1 mentions per tag post. |
| load_usertags | true | Enable or disable fetching of accessibility caption for the tag posts. Accessibility caption will be available in attached_media_content field.

+1 mentions per tag post. |
| upload_posts_image_to_s3 | true | Upload attached_media_display_url image to external data storage.

+1 credit per image. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/tag/:tag_id/update`

---

##### Cached hashtag data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/tag/:tag_id`

---

##### Cached posts from a hashtag search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/tag/:tag_id/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| sort_type | recent | Sort type for tagged posts

"recent" - open recent posts (without reels)
"top" - open top posts (including reels)
"clips" - open video posts (only reels) |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| location_name | New York | Filter items by the location name.
A query can contain quotes, logical operators, parenthesis and stopwords. |
| query | trading -forex | Filter items by the text search query.

A query can contain quotes, logical operator OR and stopwords. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### for posts by keywords

---

##### auto-update task

---

###### Search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/post/update?keywords=travel&max_posts=50&auto_update_interval=86400&auto_update_expire_at=2027-03-18T23:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | travel | List of keywords or a phrase to search |
| max_posts | 50 | Maximum number of posts to fetch.

+2 credit per post. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-03-18T23:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.

Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| callback_url | https://webhook.site/6257994e-e1c0-42d3-8b85-069dafd596bd | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/post/tasks/:task_id`

---

##### Task to update posts search by keywords

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/post/update?keywords=travel&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | travel | List of keywords or a phrase to search |
| max_posts | 50 | Maximum number of posts to fetch.

+2 credit per post. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.

Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| callback_url | https://webhook.site/b51010b8-0773-4a9d-976c-71eeef6cbf54 | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/post/update?keywords=travel`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | travel | List of keywords or a phrase to search |

---

##### Post search task details

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/post?keywords=travel`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | travel | List of keywords or a phrase to search |

---

##### Cached posts from search by keywords

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/post/items?keywords=travel&order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | travel | List of keywords or a phrase to search |
| order_by | id_desc | Enum: "id_asc" "id_desc"
Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2025-01-01 | Filter items by date (lower bound).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2025-04-30 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### for posts by location

---

##### auto-update task

---

###### Search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/location/:location_id/update?load_posts=true&max_posts=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_posts | true | Fetch tagged posts |
| max_posts | 50 | Maximum number of posts to fetch |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/location/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/instagram/location/tasks/:task_id`

---

##### Task to update location posts search

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/location/:location_id/update?load_posts=true&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_posts | true | Fetch tagged posts |
| max_posts | 50 | Maximum number of posts to fetch |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/location/:location_id/update`

---

##### Cached location data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/location/:location_id`

---

##### Cached posts from a location search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/location/:location_id/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | Filter items by date (lower bound).
All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| location_name | New York | Filter items by the location name.
A query can contain quotes, logical operators, parenthesis and stopwords. |
| query | makeup | Filter items by the text search query.

A query can contain quotes, logical operator OR and stopwords. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### music search

---

#### auto-update task

---

##### Search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/music/:music_id/update?auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/music/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MjY3NDI2NjQ0ODMxNzk4 | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/music/tasks/:task_id`

---

#### Task to update music search

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/music/:music_id/update`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/music/:music_id/update`

---

#### Music search task details

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/search/music/:music_id`

---

### post

---

#### auto-update task

---

##### Post auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/post/:post_id/update?load_comments=true&max_comments=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.
Values < 25 will drastically increase the update process performance. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_accessibility_caption | true | Enable or disable fetching of accessibility caption for the post. Accessibility caption will be available in attached_media_content field. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/instagram/post/tasks/:task_id`

---

#### Post update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/instagram/post/:post_id/update?load_comments=true&max_comments=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_accessibility_caption | true | Enable or disable fetching of accessibility caption for the post. Accessibility caption will be available in attached_media_content field. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the post task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/post/:post_id/update`

---

#### Cached post data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/post/:post_id`

---

#### Cached post comments

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/post/:post_id/comments?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached comment replies

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/comment/:comment_id/replies?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### comment

---

#### Cached post comments

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/post/:post_id/comments?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached comment data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/comment/:comment_id`

---

#### Cached comment replies to comment

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/comment/:comment_id/replies?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### Instagram queue size

**GET** `{{API_BASE_URL}}{{API_VERSION}}/instagram/queues/size`

---

### API usage stats

**GET** `{{API_BASE_URL}}{{API_VERSION}}/stats/usage?from_date=2025-03-25&to_date=2025-04-01&services=facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| from_date | 2025-03-25 | Calculate usage from this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| to_date | 2025-04-01 | Calculate usage to this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| services | facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest | Filter usage stats by services.

Multiple values can be provided as a comma-separated list. |

---

## linkedin

---

### member

---

#### auto-update task

---

##### Member auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/:member_id/update?load_activities=true&max_activities=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_activities | true | Enable or disable fetching of items from the feed of the member. |
| max_activities | 50 | Set limit for the number of items that will be fetched for the member.
We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_activity_comments | true | Enable or disable fetching of items from the comments feed of the member.

Use cached member activities created_comment_on_post activity_type to fetch results. |
| max_activity_comments | 50 | Set limit for the number of items that will be fetched for the comments feed.

We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| load_activity_reactions | true | Enable or disable fetching of items from the reactions feed of the member.

Use cached member activities reacted_to_post and reacted_to_comment_on_post activities to fetch results. |
| load_activity_reactions | 50 | Set limit for the number of items that will be fetched for the reactions feed.

We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| load_activity_articles | true | Enable or disable fetching of items from the articles feed of the member. Use cached member created_articles activity_type to fetch the results.

+1 mention per post. |
| max_activity_articles | 50 | Set limit for the number of items that will be fetched for the articles feed.

We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_about_profile | true | Load data from 'More -> About this profile' menu.

+1 mention per task. |
| load_contact_info | true | Load data from the Contact Info section of the profile.

+3 mention per task. |
| load_certifications | true | Load certifications for the profile including 'See more' page.

+1 mention per profile. |
| load_skills | true | Load skills for the profile including 'See more' page.

+1 mention per profile. |
| load_courses | true | Load courses for the profile including 'See more' page.

+1 mention per profile. |
| load_positions | true | Load positions for the profile including 'See more' page.

+1 mention per profile. |
| load_organizations | true | Load organizations for the profile including 'See more' page.

+1 mention per profile. |
| load_educations | true | Load educations for the profile including 'See more' page.

+1 mention per profile. |
| load_languages | true | Load languages for the profile including 'See more' page.

+1 mention per profile. |
| load_honors | true | Load honors for the profile including 'See more' page.

+1 mention per profile. |
| load_interests | true | Load top voices and companies from interests section of the profile including 'See more' page.

+1 mention per 20 items rounded up. |
| load_recommendations | true | Load received and given recommendations for the profile including 'See more' page.

+1 mention per 20 items rounded up. |
| load_featured_items | true | Load featured items for the profile including 'See more' page.

+1 mention per 20 items rounded up. |
| load_additional_data | true | Enable or disable fetching of data: book_an_appointment.

+3 mentions |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot.

+15 mentions per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/tasks/:task_id`

---

#### Member update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/:member_id/update?load_activities=true&max_activities=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_activities | true | Enable or disable fetching of items from the feed of the member. |
| max_activities | 50 | Set limit for the number of items that will be fetched for the member.
We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| load_activity_comments | true | Enable or disable fetching of items from the comments feed of the member.

Use cached member activities created_comment_on_post activity_type to fetch results. |
| max_activity_comments | 50 | Set limit for the number of items that will be fetched for the comments feed.

We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| load_activity_reactions | true | Enable or disable fetching of items from the reactions feed of the member.

Use cached member activities reacted_to_post and reacted_to_comment_on_post activities to fetch results. |
| max_activity_reactions | 50 | Set limit for the number of items that will be fetched for the reactions feed.

We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| load_activity_articles | true | Enable or disable fetching of items from the articles feed of the member. Use cached member created_articles activity_type to fetch the results.

+1 mention per post. |
| max_activity_articles | 50 | Set limit for the number of items that will be fetched for the articles feed.

We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_about_profile | true | Load data from 'More -> About this profile' menu.

+1 mention per task. |
| load_contact_info | true | Load data from the Contact Info section of the profile.

+3 mention per task. |
| load_certifications | true | Load certifications for the profile including 'See more' page.

+1 mention per profile. |
| load_skills | true | Load skills for the profile including 'See more' page.

+1 mention per profile. |
| load_courses | true | Load courses for the profile including 'See more' page.

+1 mention per profile. |
| load_positions | true | Load positions for the profile including 'See more' page.

+1 mention per profile. |
| load_organizations | true | Load organizations for the profile including 'See more' page.

+1 mention per profile. |
| load_educations | true | Load educations for the profile including 'See more' page.

+1 mention per profile. |
| load_languages | true | Load languages for the profile including 'See more' page.

+1 mention per profile. |
| load_honors | true | Load honors for the profile including 'See more' page.

+1 mention per profile. |
| load_interests | true | Load top voices and companies from interests section of the profile including 'See more' page.

+1 mention per 20 items rounded up. |
| load_recommendations | true | Load received and given recommendations for the profile including 'See more' page.

+1 mention per 20 items rounded up. |
| load_featured_items | true | Load featured items for the profile including 'See more' page.

+1 mention per 20 items rounded up. |
| load_additional_data | true | Enable or disable fetching of data: book_an_appointment.

+3 mentions |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot.

+15 mentions per profile. |
| callback_url | https://webhook.site/b28b3943-9c0c-41b3-9616-0c53e3d3d4fc | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the member task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/:member_id/update`

---

#### Cached member data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/:member_id`

---

#### Cached member posts

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/:member_id/activity/:activity_type/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| query | microsoft | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached member articles

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/:member_id/activity/:activity_type/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| query | microsoft | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached posts commented by the member

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/:member_id/activity/:activity_type/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| query | microsoft | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached posts reacted by the member

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/:member_id/activity/:activity_type/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| query | microsoft | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### C.p. where the member reacted to a comment

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/:member_id/activity/:activity_type/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| query | microsoft | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### company

---

#### auto-update task

---

##### Company auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/:company_id/update?load_feed_posts=true&max_feed_posts=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_feed_posts | true | Enable or disable fetching of posts from the feed of the company. |
| max_feed_posts | 50 | Set limit for the number of posts that will be fetched for the company.
We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_video_posts | true | Enable or disable fetching of posts from the videos section of the company. |
| max_video_posts | 50 | Set limit for the number of posts from the videos section that will be fetched for the company.

We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_alumni_count | true | Enable or disable fetching of alumni count for schools.

+1 mention per company. |
| load_affiliated_companies | true | Enable or disable fetching of affiliated companies for parent company.

+1 mention per one fetched affiliated company. |
| load_people_data | true | Enable fetching of data from People tab of the company.

+10 mentions |
| load_additional_data | true | Enable or disable fetching of data: is_claimable, highlight_items, parent_company, is_auto_generated, acquired_by .

+3 mentions |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot. School pages are not supported yet. +15 credits per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/tasks/:task_id`

---

#### Company update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/:company_id/update?load_feed_posts=true&max_feed_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_feed_posts | true | Enable or disable fetching of posts from the feed of the company. |
| max_feed_posts | 50 | Set limit for the number of posts that will be fetched for the company.
We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| load_video_posts | true | Enable or disable fetching of posts from the videos section of the company. |
| max_video_posts | 50 | Set limit for the number of posts from the videos section that will be fetched for the company.

We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_alumni_count | true | Enable or disable fetching of alumni count for schools.

+1 mention per company. |
| load_affiliated_companies | true | Enable or disable fetching of affiliated companies for parent company.

+1 mention per one fetched affiliated company. |
| load_people_data | true | Enable fetching of data from People tab of the company.

+10 mentions |
| load_additional_data | true | Enable or disable fetching of data: is_claimable, highlight_items, parent_company, is_auto_generated, acquired_by .

+3 credits |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot. School pages are not supported yet. +15 credits per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the company task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/:company_id/update`

---

#### Cached company`s data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/:company_id`

---

#### Cached company`s posts

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/:company_id/feed/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| query | -"Satya Nadella" | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached company`s video posts

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/:company_id/video/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| query | -"Satya Nadella" | Filter items by the text search query.
A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | ZGF0ZV9kZXNjfDIwMjItMDItMDEgMDA6MDA6MDA= | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### search

---

#### for posts

---

##### auto-update task

---

###### Posts search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/search/update?keywords=italy&content_type=videos&sort_type=most_recent&date_posted=past_24h&max_results=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | italy | Enables "Keywords" filter when fetching search results. You can use title, skill, or company as values. |
| content_type | videos | Enables "Content Type" filter when fetching search results. |
| sort_type | most_recent | Enables "Sort By" filter when fetching search results. |
| date_posted | past_24h | Enables "Date Posted" filter when fetching search results. |
| max_results | 50 | Set limit for the number of results that will be fetched for the search. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| from_date | 2022-01-01 | Filter items by date (lower bound).

Timestamp value must be in ISO 8601 format. - Items are collected from newest to oldest and the update process will stop once the specified from_date is reached. - If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run. |
| companies | 1035,1586 | Enables "Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target companies' IDs. |
| members | ACoAADNfMCMBK6LHFLe4f4_BTiJP1by_iakUzCQ,ACoAAAJhZBwBBioJ8YnqkZgr6JTP6L8g71IdYiA | Enables "From member" filter when fetching search results.
The parameter must set to a comma-separated list of the target members fsd_profile IDs.
We recommend to use no more than 10 members per request or update task may fail. |
| from_companies | 1035 | Enables "From company" filter when fetching search results.

The parameter must set to a comma-separated list of the target target companies' IDs. |
| mentioning_members | ACoAAAAK870Bd6yR_NZfTZeo7zFQd61oPbdwhCY | Enables "Mentioning member" filter when fetching search results.

The parameter must set to a comma-separated list of the target members fsd_profile IDs. |
| mentioning_companies | 1035 | Enables "Mentioning company" filter when fetching search results.

The parameter must set to a comma-separated list of the target companies' IDs. |
| author_industries | 1035 | Enables "Author industry" filter when fetching search results.

The parameter must set to a comma-separated list of the target industries IDs. |
| author_keywords_title | HR | Enables "Author Keywords - Title" filter when fetching search results. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/search/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/search/tasks/:task_id`

---

##### Posts search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/search/update?keywords=italy&content_type=videos&sort_type=most_recent&date_posted=past_24h&max_results=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | italy | Enables "Keywords" filter when fetching search results. You can use title, skill, or company as values. |
| content_type | videos | Enables "Content Type" filter when fetching search results. |
| sort_type | most_recent | Enables "Sort By" filter when fetching search results. |
| date_posted | past_24h | Enables "Date Posted" filter when fetching search results. |
| max_results | 50 | Set limit for the number of results that will be fetched for the search. |
| from_date | 2022-01-01 | Filter items by date (lower bound).

Timestamp value must be in ISO 8601 format. - Items are collected from newest to oldest and the update process will stop once the specified from_date is reached. - If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run. |
| companies | 1035,1586 | Enables "Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target companies' IDs. |
| members | ACoAADNfMCMBK6LHFLe4f4_BTiJP1by_iakUzCQ,ACoAAAJhZBwBBioJ8YnqkZgr6JTP6L8g71IdYiA | Enables "From member" filter when fetching search results.

The parameter must set to a comma-separated list of the target members fsd_profile IDs.
We recommend to use no more than 10 members per request or update task may fail. |
| from_companies | 1035 | Enables "From company" filter when fetching search results.

The parameter must set to a comma-separated list of the target target companies' IDs. |
| mentioning_members | ACoAAAAK870Bd6yR_NZfTZeo7zFQd61oPbdwhCY | Enables "Mentioning member" filter when fetching search results.

The parameter must set to a comma-separated list of the target members fsd_profile IDs. |
| mentioning_companies | 1035 | Enables "Mentioning company" filter when fetching search results.

The parameter must set to a comma-separated list of the target companies' IDs. |
| author_industries | 1035 | Enables "Author industry" filter when fetching search results.

The parameter must set to a comma-separated list of the target industries IDs. |
| author_keywords_title | HR | Enables "Author Keywords - Title" filter when fetching search results. |
| load_comments | true | Enable or disable fetching of comments for every post. |
| max_comments | 50 | Set limits for the number of comments that will be fetched for every post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/search/update?keywords=italy&content_type=videos&sort_type=most_recent&date_posted=past_24h`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | italy | Enables "Keywords" filter when fetching search results. You can use title, skill, or company as values. |
| content_type | videos | Enables "Content Type" filter when fetching search results. |
| sort_type | most_recent | Enables "Sort By" filter when fetching search results. |
| date_posted | past_24h | Enables "Date Posted" filter when fetching search results. |
| companies | 1035,1586 | Enables "Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target companies' IDs. |
| members | ACoAADNfMCMBK6LHFLe4f4_BTiJP1by_iakUzCQ,ACoAAAJhZBwBBioJ8YnqkZgr6JTP6L8g71IdYiA | Enables "From member" filter when fetching search results.

The parameter must set to a comma-separated list of the target members fsd_profile IDs.
We recommend to use no more than 10 members per request or update task may fail. |
| from_companies | 1035 | Enables "From company" filter when fetching search results.

The parameter must set to a comma-separated list of the target target companies' IDs. |
| mentioning_members | ACoAAAAK870Bd6yR_NZfTZeo7zFQd61oPbdwhCY | Enables "Mentioning member" filter when fetching search results.

The parameter must set to a comma-separated list of the target members fsd_profile IDs. |
| mentioning_companies | 1035 | Enables "Mentioning company" filter when fetching search results.

The parameter must set to a comma-separated list of the target companies' IDs. |
| author_industries | 1035 | Enables "Author industry" filter when fetching search results.

The parameter must set to a comma-separated list of the target industries IDs. |
| author_keywords_title | HR | Enables "Author Keywords - Title" filter when fetching search results. |

---

##### Cached posts from posts search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/search/posts?keywords=italy&content_type=videos&sort_type=most_recent&date_posted=past_24h&order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | italy | Enables "Keywords" filter when fetching search results. You can use title, skill, or company as values. |
| content_type | videos | Enables "Content Type" filter when fetching search results. |
| sort_type | most_recent | Enables "Sort By" filter when fetching search results. |
| date_posted | past_24h | Enables "Date Posted" filter when fetching search results. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| companies | 1035,1586 | Enables "Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target companies' IDs. |
| members | ACoAADNfMCMBK6LHFLe4f4_BTiJP1by_iakUzCQ,ACoAAAJhZBwBBioJ8YnqkZgr6JTP6L8g71IdYiA | Enables "From member" filter when fetching search results.
The parameter must set to a comma-separated list of the target members fsd_profile IDs.
We recommend to use no more than 10 members per request or update task may fail. |
| from_companies | 1035 | Enables "From company" filter when fetching search results.

The parameter must set to a comma-separated list of the target target companies' IDs. |
| mentioning_members | ACoAAAAK870Bd6yR_NZfTZeo7zFQd61oPbdwhCY | Enables "Mentioning member" filter when fetching search results.

The parameter must set to a comma-separated list of the target members fsd_profile IDs. |
| mentioning_companies | 1035 | Enables "Mentioning company" filter when fetching search results.

The parameter must set to a comma-separated list of the target companies' IDs. |
| author_industries | 1035 | Enables "Author industry" filter when fetching search results.

The parameter must set to a comma-separated list of the target industries IDs. |
| author_keywords_title | HR | Enables "Author Keywords - Title" filter when fetching search results. |
| from_date | 2021-01-01 | Filter items by date (lower bound).
All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2022-01-01 | Filter items by date (upper bound).
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### for jobs

---

##### auto-update task

---

###### Jobs search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/job/search/update?keywords=developer&sort_type=most_recent&date_posted=past_24h&max_results=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | developer | Enables "Keywords" filter when fetching search results. You can use title, skill, or company as values. |
| sort_type | most_recent | Enables "Sort By" filter when fetching search results. |
| date_posted | past_24h | Enables "Date Posted" filter when fetching search results. |
| max_results | 50 | Set limit for the number of results that will be fetched for the search. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| location | European Union | Enables "Location" filter when fetching search results. You can use region ("European Union"), country, city, state, or zip code as values. Default value is "worldwide". |
| companies | 1035,1586 | Enables "Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target companies' IDs. |
| experience_levels | mid_senior_level,entry_level,associate | Enables "Experience Level" filter when fetching search results.
The parameter must set to a comma-separated list of the experience levels' names. |
| job_types | full_time,part_time | Enables "Job Type" filter when fetching search results.
The parameter must set to a comma-separated list of the job types' names. |
| remote | true | Enables "Remote" filter when fetching search results. |
| under_10_applicants | true | Enables "Under 10 Applicants" filter when fetching search results. |
| titles | true | The parameter must set to a comma-separated list of the target job titles ids. |
| load_skill_details | true | Enables loading of "Skills Details" when fetching job results.

+1 mention per job post |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/job/search/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/job/search/tasks/:task_id`

---

##### Jobs search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/job/search/update?keywords=developer&sort_type=most_recent&date_posted=past_24h&max_results=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | developer | Enables "Keywords" filter when fetching search results. You can use title, skill, or company as values. |
| sort_type | most_recent | Enables "Sort By" filter when fetching search results. |
| date_posted | past_24h | Enables "Date Posted" filter when fetching search results. |
| max_results | 50 | Set limit for the number of results that will be fetched for the search. |
| location | European Union | Enables "Location" filter when fetching search results. You can use region ("European Union"), country, city, state, or zip code as values. Default value is "worldwide". |
| companies | 1035,1586 | Enables "Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target companies' IDs. |
| experience_levels | mid_senior_level,entry_level,associate | Enables "Experience Level" filter when fetching search results.
The parameter must set to a comma-separated list of the experience levels' names. |
| job_types | full_time,part_time | Enables "Job Type" filter when fetching search results.
The parameter must set to a comma-separated list of the job types' names. |
| remote | true | Enables "Remote" filter when fetching search results. |
| under_10_applicants | true | Enables "Under 10 Applicants" filter when fetching search results. |
| titles | true | The parameter must set to a comma-separated list of the target job titles ids. |
| load_skill_details | true | Enables loading of "Skills Details" when fetching job results.

+1 mention per job post |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/job/search/update?keywords=developer&sort_type=most_recent&date_posted=past_24h`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | developer | Enables "Keywords" filter when fetching search results. You can use title, skill, or company as values. |
| sort_type | most_recent | Enables "Sort By" filter when fetching search results. |
| date_posted | past_24h | Enables "Date Posted" filter when fetching search results. |
| location | European Union | Enables "Location" filter when fetching search results. You can use region ("European Union"), country, city, state, or zip code as values. Default value is "worldwide". |
| companies | 1035,1586 | Enables "Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target companies' IDs. |
| experience_levels | mid_senior_level,entry_level,associate | Enables "Experience Level" filter when fetching search results.
The parameter must set to a comma-separated list of the experience levels' names. |
| job_types | full_time,part_time | Enables "Job Type" filter when fetching search results.
The parameter must set to a comma-separated list of the job types' names. |
| remote | 1 | Enables "Remote" filter when fetching search results. |
| under_10_applicants | 0 | Enables "Under 10 Applicants" filter when fetching search results. |
| titles | true | The parameter must set to a comma-separated list of the target job titles ids. |

---

##### Cached items from jobs search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/job/search/posts?keywords=developer&sort_type=most_recent&date_posted=past_24h&order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | developer | Enables "Keywords" filter when fetching search results. You can use title, skill, or company as values. |
| sort_type | most_recent | Enables "Sort By" filter when fetching search results. |
| date_posted | past_24h | Enables "Date Posted" filter when fetching search results. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| location | European Union | Enables "Location" filter when fetching search results. You can use region ("European Union"), country, city, state, or zip code as values. Default value is "worldwide". |
| companies | 1035,1586 | Enables "Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target companies' IDs. |
| experience_levels | mid_senior_level,entry_level,associate | Enables "Experience Level" filter when fetching search results.
The parameter must set to a comma-separated list of the experience levels' names. |
| job_types | full_time,part_time | Enables "Job Type" filter when fetching search results.
The parameter must set to a comma-separated list of the job types' names. |
| remote | 1 | Enables "Remote" filter when fetching search results. |
| under_10_applicants | 0 | Enables "Under 10 Applicants" filter when fetching search results. |
| titles | true | The parameter must set to a comma-separated list of the target job titles ids. |
| from_date | 2021-01-01 | Filter items by date (lower bound).
All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### for members

---

##### auto-update task

---

###### Members search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/search/update?keywords=John Murray&max_results=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | John Murray | Enables "Keywords" filter when fetching search results. You can use title, skill, or company as values. |
| max_results | 50 | Set limit for the number of results that will be fetched for the search. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| keywords_title | IT Project Manager  | Enables "Keywords - Title" filter when fetching search results. |
| keywords_first_name | Rostislav | Enables "Keywords - First Name" filter when fetching search results. |
| keywords_last_name | Dmitrenko | Enables "Keywords - Last Name" filter when fetching search results. |
| keywords_school | 46 Kharkiv | Enables "Keywords - School" filter when fetching search results. |
| keywords_company | Reikartz Hotel Group | Enables "Keywords - Company" filter when fetching search results. |
| companies | 3570 | Enables "Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target companies' IDs. |
| past_companies | 1035,1586 | Enables "Past Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target past companies' IDs. |
| locations | 103644278 | Enables "Locations" filter when fetching search results.
The parameter must set to a comma-separated list of the target locations' IDs. |
| industries | 96 | Enables "Industries" filter when fetching search results.
The parameter must set to a comma-separated list of the target industries' IDs. |
| languages | en | Enables "Languages" filter when fetching search results.
The parameter must set to a comma-separated list of the target languages'. |
| service_categories | 220 | Enables "Service Categories" filter when fetching search results.
The parameter must set to a comma-separated list of the target service categories' IDs. |
| actively_hiring | true | Enables "Actively Hiring" filter when fetching search results.

+100 mentions |
| load_profiles_data | true | Fetch all fields for profiles.
If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/search/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/search/tasks/:task_id`

---

##### Members search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/search/update?keywords=Ukraine&max_results=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | Ukraine | Enables "Keywords" filter when fetching search results. You can use title, skill, or company as values. |
| max_results | 50 | Set limit for the number of results that will be fetched for the search. |
| keywords_title | IT Project Manager  | Enables "Keywords - Title" filter when fetching search results. |
| keywords_first_name | Rostislav | Enables "Keywords - First Name" filter when fetching search results. |
| keywords_last_name | Dmitrenko | Enables "Keywords - Last Name" filter when fetching search results. |
| keywords_company | Reikartz Hotel Group | Enables "Keywords - Company" filter when fetching search results. |
| keywords_school | 46 Kharkiv | Enables "Keywords - School" filter when fetching search results. |
| companies | 3570 | Enables "Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target companies' IDs. |
| past_companies | 1035,1586 | Enables "Past Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target past companies' IDs. |
| locations | 103644278 | Enables "Locations" filter when fetching search results.
The parameter must set to a comma-separated list of the target locations' IDs. |
| industries | 96 | Enables "Industries" filter when fetching search results.
The parameter must set to a comma-separated list of the target industries' IDs. |
| languages | en | Enables "Languages" filter when fetching search results.
The parameter must set to a comma-separated list of the target languages'. |
| service_categories | 220 | Enables "Service Categories" filter when fetching search results.
The parameter must set to a comma-separated list of the target service categories' IDs. |
| actively_hiring | true | Enables "Actively Hiring" filter when fetching search results.

+100 mentions |
| load_profiles_data | true | Fetch all fields for profiles.

If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.

+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/search/update?keywords=Ukraine`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | Ukraine | Enables "Keywords" filter when fetching search results. You can use title, skill, or company as values. |
| keywords_title | IT Project Manager  | Enables "Keywords - Title" filter when fetching search results. |
| keywords_first_name | Rostislav | Enables "Keywords - First Name" filter when fetching search results. |
| keywords_last_name | Dmitrenko | Enables "Keywords - Last Name" filter when fetching search results. |
| keywords_company | Reikartz Hotel Group | Enables "Keywords - Company" filter when fetching search results. |
| keywords_school | 46 Kharkiv | Enables "Keywords - School" filter when fetching search results. |
| companies | 3570 | Enables "Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target companies' IDs. |
| past_companies | 1035,1586 | Enables "Past Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target past companies' IDs. |
| locations | 103644278 | Enables "Locations" filter when fetching search results.
The parameter must set to a comma-separated list of the target locations' IDs. |
| industries | 96 | Enables "Industries" filter when fetching search results.
The parameter must set to a comma-separated list of the target industries' IDs. |
| languages | en | Enables "Languages" filter when fetching search results.
The parameter must set to a comma-separated list of the target languages'. |
| service_categories | 220 | Enables "Service Categories" filter when fetching search results.
The parameter must set to a comma-separated list of the target service categories' IDs. |
| actively_hiring | true | Enables "Actively Hiring" filter when fetching search results.

+100 mentions |

---

##### Cached members from members search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/member/search/members?keywords=Ukraine&order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | Ukraine | Enables "Keywords" filter when fetching search results. You can use title, skill, or company as values. |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| keywords_title | IT Project Manager  | Enables "Keywords - Title" filter when fetching search results. |
| keywords_first_name | Rostislav | Enables "Keywords - First Name" filter when fetching search results. |
| keywords_last_name | Dmitrenko | Enables "Keywords - Last Name" filter when fetching search results. |
| keywords_company | Reikartz Hotel Group | Enables "Keywords - Company" filter when fetching search results. |
| keywords_school | 46 Kharkiv | Enables "Keywords - School" filter when fetching search results. |
| companies | 3570 | Enables "Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target companies' IDs. |
| past_companies | 1035,1586 | Enables "Past Company" filter when fetching search results.
The parameter must set to a comma-separated list of the target past companies' IDs. |
| locations | 103644278 | Enables "Locations" filter when fetching search results.
The parameter must set to a comma-separated list of the target locations' IDs. |
| industries | 96 | Enables "Industries" filter when fetching search results.
The parameter must set to a comma-separated list of the target industries' IDs. |
| languages | en | Enables "Languages" filter when fetching search results.
The parameter must set to a comma-separated list of the target languages'. |
| service_categories | 220 | Enables "Service Categories" filter when fetching search results.
The parameter must set to a comma-separated list of the target service categories' IDs. |
| actively_hiring | true | Enables "Actively Hiring" filter when fetching search results.

+100 mentions |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor).Enables "Talks About" filter when fetching search results.
The parameter must set to a comma-separated list of the target hashtags. |

---

#### for companies

---

##### auto-update task

---

###### Companies search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/search/update?keywords=Amazon&max_results=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | Amazon | Enables "Keywords" filter when fetching search results. |
| max_results | 50 | Set limit for the number of results that will be fetched for the search. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| locations | 103644278 | Enables "Locations" filter when fetching search results.
The parameter must set to a comma-separated list of the target locations' IDs. |
| industries | 96 | Enables "Industries" filter when fetching search results.
The parameter must set to a comma-separated list of the target industries' IDs. |
| sizes | B | Example: sizes=B,C
Enables "Company size" filter when fetching search results.

The parameter must set to a comma-separated list of the target company size IDs.
Supported values:

B -> 1-10 employees
C -> 11-50 employees
D -> 51-200 employees
E -> 201-500 employees
F -> 501-1000 employees
G -> 1001-5000 employees
H -> 5001-10,000 employees
I -> 10,001+ employees |
| load_profiles_data | true | Fetch all fields for profiles.
If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/search/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/search/tasks/:task_id`

---

##### Companies search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/search/update?keywords=Amazon&max_results=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | Amazon | Enables "Keywords" filter when fetching search results. |
| max_results | 50 | Set limit for the number of results that will be fetched for the search. |
| locations | 103644278 | Enables "Locations" filter when fetching search results.
The parameter must set to a comma-separated list of the target locations' IDs. |
| industries | 96 | Enables "Industries" filter when fetching search results.
The parameter must set to a comma-separated list of the target industries' IDs. |
| sizes | B | Example: sizes=B,C
Enables "Company size" filter when fetching search results.

The parameter must set to a comma-separated list of the target company size IDs.
Supported values:

B -> 1-10 employees
C -> 11-50 employees
D -> 51-200 employees
E -> 201-500 employees
F -> 501-1000 employees
G -> 1001-5000 employees
H -> 5001-10,000 employees
I -> 10,001+ employees |
| load_profiles_data | true | Fetch all fields for profiles.

If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.

+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/search/update?keywords=Amazon`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | Amazon | Enables "Keywords" filter when fetching search results. |
| locations | 103644278 | Enables "Locations" filter when fetching search results.
The parameter must set to a comma-separated list of the target locations' IDs. |
| industries | 96 | Enables "Industries" filter when fetching search results.
The parameter must set to a comma-separated list of the target industries' IDs. |
| sizes | B | Example: sizes=B,C
Enables "Company size" filter when fetching search results.

The parameter must set to a comma-separated list of the target company size IDs.
Supported values:

B -> 1-10 employees
C -> 11-50 employees
D -> 51-200 employees
E -> 201-500 employees
F -> 501-1000 employees
G -> 1001-5000 employees
H -> 5001-10,000 employees
I -> 10,001+ employees |

---

##### Cached companies from companies search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/company/search/items?keywords=Amazon&order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | Amazon | Enables "Keywords" filter when fetching search results. |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| locations | 103644278 | Enables "Locations" filter when fetching search results.
The parameter must set to a comma-separated list of the target locations' IDs. |
| industries | 96 | Enables "Industries" filter when fetching search results.
The parameter must set to a comma-separated list of the target industries' IDs. |
| sizes | B | Example: sizes=B,C
Enables "Company size" filter when fetching search results.

The parameter must set to a comma-separated list of the target company size IDs.
Supported values:

B -> 1-10 employees
C -> 11-50 employees
D -> 51-200 employees
E -> 201-500 employees
F -> 501-1000 employees
G -> 1001-5000 employees
H -> 5001-10,000 employees
I -> 10,001+ employees |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### for ad posts

---

##### auto-update task

---

###### Ad posts search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/ad-post/search/update?from_companies=1035&max_results=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| from_companies | 1035 | Fetch search results from the following company. |
| max_results | 50 | Set limit for the number of results that will be fetched for the search. When max_results=0 is set, you will get the total number of found results |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/ad-post/search/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/ad-post/search/tasks/:task_id`

---

##### Ad posts search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/ad-post/search/update?from_companies=1035&max_results=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| from_companies | 1035 | Fetch search results from the following company. |
| max_results | 50 | Set limit for the number of results that will be fetched for the search. When max_results=0 is set, you will get the total number of found results. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/ad-post/search/update?from_companies=1035`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| from_companies | 1035 | Fetch search results from the following company. |

---

##### Cached posts from ad posts search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/ad-post/search/items?from_companies=1035&order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| from_companies | 1035 | Fetch search results from the following company. |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### post

---

#### auto-update task

---

##### Post auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/:post_id/update?load_comments=true&max_comments=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.
We do not recommend setting this parameter to values > 100 as it can cause the update process to fail.
Values < 25 will drastically increase the update process performance. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_reactors | true | Enable or disable fetching of reactors for the post. |
| max_reactors | 50 | Set limit for the number of reactors that will be fetched for the post. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/tasks/:task_id`

---

#### Post update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/:post_id/update?load_comments=true&max_comments=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.
We do not recommend setting this parameter to values > 100 as it can cause the update process to fail.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_reactors | true | Enable or disable fetching of reactors for the post. |
| max_reactors | 50 | Set limit for the number of reactors that will be fetched for the post. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the post update task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/:post_id/update`

---

#### Cached post data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/:post_id`

---

#### Cached post comments

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/:post_id/comments?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | Filter items by date (lower bound).
All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached comment replies

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/comment/:comment_id/replies?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | Filter items by date (lower bound).
All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached post reactors (like) member profiles

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/:post_id/reactors/:reaction_type/members?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached post reactors (like) company profiles

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/:post_id/reactors/:reaction_type/companies?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### comment

---

#### Cached post comments

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/post/:post_id/comments?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | Filter items by date (lower bound).
All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached comment data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/comment/:comment_id`

---

#### Cached comment replies

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/comment/:comment_id/replies?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | Filter items by date (lower bound).
All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).
All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### Linkedin queue size

**GET** `{{API_BASE_URL}}{{API_VERSION}}/linkedin/queues/size`

---

### API usage stats

**GET** `{{API_BASE_URL}}{{API_VERSION}}/stats/usage?from_date=2025-03-25&to_date=2025-04-01&services=facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| from_date | 2025-03-25 | Calculate usage from this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| to_date | 2025-04-01 | Calculate usage to this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| services | facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest | Filter usage stats by services.

Multiple values can be provided as a comma-separated list. |

---

## tiktok

---

### profile

---

#### auto-update task

---

##### Profile auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id/update?load_feed_posts=true&max_posts=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_feed_posts | true | Enable or disable fetching of posts from the feed section of the profile. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the profile. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| from_date | 2022-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_followers | true | Enable or disable fetching of followers of the profile.

+9 credit per profile info |
| load_followers_data | true | This parameter can be enabled only if load_followers parameter is enabled.
+9 credits per follower. |
| max_followers | 50 | Set a limit for the number of followers that will be fetched for the profile. |
| load_following | true | Enable or disable fetching of following of the profile.

+9 credit per profile info |
| load_following_data | true | This parameter can be enabled only if load_following parameter is enabled.
+9 credits per follower. |
| max_following | 50 | Set a limit for the number of following that will be fetched for the profile. |
| page_screenshot | true | Take full page screenshot.

+15 credits per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/tasks/:task_id`

---

#### Profile update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id/update?load_feed_posts=true&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_feed_posts | true | Enable or disable fetching of posts from the feed section of the profile. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the profile. |
| from_date | 2022-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.

Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| load_followers | true | Enable or disable fetching of followers of the profile.

+9 credit per profile info |
| load_followers_data | true | This parameter can be enabled only if load_followers parameter is enabled.
+9 credits per follower. |
| max_followers | 50 | Set a limit for the number of followers that will be fetched for the profile. |
| load_following | true | Enable or disable fetching of following of the profile.

+9 credit per profile info |
| load_following_data | true | This parameter can be enabled only if load_following parameter is enabled.
+9 credits per follower. |
| max_following | 50 | Set a limit for the number of following that will be fetched for the profile. |
| page_screenshot | true | Take full page screenshot.

+15 credits per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the profile task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id/update`

---

#### Cached profile data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id`

---

#### Cached profile posts

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id/feed/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-06-01 | Filter items by date (lower bound).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format.
Supports integer values to calculate relative datetime from UTC now()

Examples:
from_date=2021-04-06T21:42:12 - Date and time
from_date=2021-04-06 - Date only (time will be set to 00:00)
from_date=-86400 - Negative integer (datetime will be set to `now()-24h`) |
| to_date | 2021-09-01 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format.
Supports integer values to calculate relative datetime from UTC now()

Examples:
to_date=2021-04-06T21:42:12 - Date and time
to_date=2021-04-06 - Date only (time will be set to 00:00)
to_date=-86400 - Negative integer (datetime will be set to `now()-24h`) |
| query | song | Filter items by the text search query.

A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| lang | en | Filter items by text language.
Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| cursor | MjY3NDI2NjQ0ODMxNzk4 | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached profile followers profiles

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id/followers/profiles?order_by=id_desc&max_page_size=100`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 100 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MjY3NDI2NjQ0ODMxNzk4 | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached profile following profiles

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id/following/profiles?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MjY3NDI2NjQ0ODMxNzk4 | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

### post

---

#### auto-update task

---

##### Post auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id/feed/posts/:post_id/update?auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| upload_video_cover_url_s3 | true | Upload video cover image to external data storage.

+1 credit per image. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page. [1 .. 100] |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/post/tasks/:task_id`

---

#### Post update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id/feed/posts/:post_id/update?load_comments=true&max_comments=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| upload_video_cover_url_s3 | true | Upload video cover image to external data storage.

+1 credit per image. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the post task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id/feed/posts/:post_id/update`

---

#### Cached post data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id/feed/posts/:post_id`

---

#### Cached post comments

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id/feed/posts/:post_id/comments?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached comment replies

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/comment/:comment_id/replies?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### posts search

---

#### for top posts by keyword

---

##### auto-update task

---

###### Post search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/post/update?keywords=world cup&load_posts=true&max_posts=50&sort_type=relevance&date_posted=all&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | world cup | List of keywords or a phrase to search |
| load_posts | true | Enable or disable fetching of posts from the search results. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the search. |
| sort_type | relevance | Enum: "relevance" "like_count" "date_posted"
Sort type for the posts. |
| date_posted | all | Enum: "all" "past_day" "past_week" "past_month" "past_half_year"
Date posted for the posts. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| upload_video_cover_url_s3 | true | Upload video cover image to external data storage.

+1 credit per image. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/post/tasks/:task_id`

---

##### Post search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/post/update?keywords=world cup&load_posts=true&max_posts=50&sort_type=relevance&date_posted=all`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | world cup | List of keywords or a phrase to search |
| load_posts | true | Enable or disable fetching of posts from the search results. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the search. |
| sort_type | relevance | Enum: "relevance" "like_count" "date_posted"
Sort type for the posts. |
| date_posted | all | Enum: "all" "past_day" "past_week" "past_month" "past_half_year"
Date posted for the posts. |
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.

Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| upload_video_cover_url_s3 | true | Upload video cover image to external data storage.

+1 credit per image. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the  search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/post/update?keywords=world cup&sort_type=relevance&date_posted=all`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | world cup | List of keywords or a phrase to search |
| sort_type | relevance | Enum: "relevance" "like_count" "date_posted"
Sort type for the posts. |
| date_posted | all | Enum: "all" "past_day" "past_week" "past_month" "past_half_year"
Date posted for the posts. |

---

##### Cached posts from posts search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/post/items?keywords=world cup&order_by=date_desc&max_page_size=50&sort_type=relevance&date_posted=all`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | world cup | List of keywords or a phrase to search

You can use logical operators (AND, OR, etc) and stopwords. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| sort_type | relevance | Enum: "relevance" "like_count" "date_posted"
Sort type for the posts. |
| date_posted | all | Enum: "all" "past_day" "past_week" "past_month" "past_half_year"
Date posted for the posts. |
| from_date | 2021-01-01 | Filter items by date (lower bound).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

#### for top posts by hashtag

---

##### auto-update task

---

###### Hashtag search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/hashtag/:hashtag_id/update?load_posts=true&max_posts=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_posts | true | Enable or disable fetching of posts from the search results. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the search. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| upload_video_cover_url_s3 | true | Upload video cover image to external data storage.

+1 credit per image |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/hashtag/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/hashtag/tasks/:task_id`

---

##### Hashtag search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/hashtag/:hashtag_id/update?load_posts=true&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_posts | true | Enable or disable fetching of posts from the search results. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the search. |
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.

Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| upload_video_cover_url_s3 | true | Upload video cover image to external data storage.

+1 credit per image. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the hashtag search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/hashtag/:hashtag_id/update`

---

##### Search task details

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/hashtag/:hashtag_id`

---

##### Cached posts from a hashtag search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/hashtag/:hashtag_id/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | Filter items by date (lower bound).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format.
Supports integer values to calculate relative datetime from UTC now()

Examples:
from_date=2021-04-06T21:42:12 - Date and time
from_date=2021-04-06 - Date only (time will be set to 00:00)
from_date=-86400 - Negative integer (datetime will be set to `now()-24h`) |
| to_date | 2023-01-01 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format.
Supports integer values to calculate relative datetime from UTC now()

Examples:
to_date=2021-04-06T21:42:12 - Date and time
to_date=2021-04-06 - Date only (time will be set to 00:00)
to_date=-86400 - Negative integer (datetime will be set to `now()-24h`) |
| query | song | Filter items by the text search query.

A query can contain quotes, logical operator OR and stopwords. |
| lang | en | Filter items by text language.

Language must be specified as two-letter ISO 639-1 code.

Supports multiple comma-separated values. |
| cursor | MjY3NDI2NjQ0ODMxNzk4 | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

#### for top posts by music

---

##### auto-update task

---

###### Music search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/music/:music_title/:music_id/update?load_posts=true&max_posts=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_posts | true | Enable or disable fetching of posts from the search results. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the search. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.
Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/music/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/music/tasks/:task_id`

---

##### Music search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/music/:music_title/:music_id/update?load_posts=true&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_posts | true | Enable or disable fetching of posts from the search results. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the search. |
| load_comments | true | Enable or disable fetching of comments for the post. |
| max_comments | 50 | Set limit for the number of comments that will be fetched for the post.

Values < 25 will drastically increase the update process performance. |
| load_replies | true | Enable or disable fetching of replies for the comments. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/music/:music_title/:music_id/update`

---

##### Search task details

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/music/:music_title/:music_id`

---

##### Cached posts from a music search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/music/:music_title/:music_id/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | Filter items by date (lower bound).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format.
Supports integer values to calculate relative datetime from UTC now()

Examples:
from_date=2021-04-06T21:42:12 - Date and time
from_date=2021-04-06 - Date only (time will be set to 00:00)
from_date=-86400 - Negative integer (datetime will be set to `now()-24h`) |
| to_date | 2023-01-01 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format.
Supports integer values to calculate relative datetime from UTC now()

Examples:
to_date=2021-04-06T21:42:12 - Date and time
to_date=2021-04-06 - Date only (time will be set to 00:00)
to_date=-86400 - Negative integer (datetime will be set to `now()-24h`) |
| query | song | Filter items by the text search query.

A query can contain quotes, logical operator OR and stopwords. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |
| lang | en | Filter items by text language.

Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |

---

### profile search

---

#### auto-update task

---

##### Search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/profile/update?keywords=bill &max_profiles=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | bill  | List of keywords or a phrase to search |
| max_profiles | 50 | Maximum number of profiles to fetch |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.
API will make a POST request to the callback URL every time the update process is finished.
This parameter is useful if you need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.
+1 credit per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/profile/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.
"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/profile/tasks/:task_id`

---

#### Profiles search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/profile/update?keywords=bill &max_profiles=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | bill  | List of keywords or a phrase to search |
| max_profiles | 50 | Maximum number of profiles to fetch |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.

+1 credit per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.
API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/profile/update?keywords=bill `

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | bill  | List of keywords or a phrase to search |

---

#### Cached profiles from a profile search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/search/profile/items?keywords=bill &order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | bill  | List of keywords or a phrase to search |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### comment

---

#### Cached post comments

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/profile/:profile_id/feed/posts/:post_id/comments?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached comment data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/comment/:comment_id`

---

#### Cached comment replies

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/comment/:comment_id/replies?max_page_size=50&order_by=date_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| from_date | 2021-01-01 | All items created before this date will be removed from the response. |
| to_date | 2023-01-01 | All items created after and at this date will be removed from the response. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### Tiktok queue size

**GET** `{{API_BASE_URL}}{{API_VERSION}}/tiktok/queues/size`

---

### API usage stats

**GET** `{{API_BASE_URL}}{{API_VERSION}}/stats/usage?from_date=2025-03-25&to_date=2025-04-01&services=facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| from_date | 2025-03-25 | Calculate usage from this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| to_date | 2025-04-01 | Calculate usage to this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| services | facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest | Filter usage stats by services.

Multiple values can be provided as a comma-separated list. |

---

## twitter

---

### profile

---

#### auto-update task

---

##### Profile auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/update?load_feed_posts=true&max_posts=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_feed_posts | true | Enable or disable fetching of posts from the feed section of the profile. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the profile. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| from_date | 2021-01-01 | Filter items by date (lower bound).

Timestamp value must be in ISO 8601 format. - Items are collected from newest to oldest and the update process will stop once the specified from_date is reached. - If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run. |
| load_feed_posts_replies | true | Enable or disable fetching of posts replies from the feed section of the profile. |
| max_feed_posts_replies | 50 | Set limit for the number of posts replies that will be fetched for the profile. |
| load_followers | true | Enable or disable fetching of followers of the profile.

+3 credit per follower. |
| max_followers | 50 | Set a limit for the number of followers that will be fetched for the profile. |
| load_following | true | Enable or disable fetching of following of the profile.

+3 credit per following. |
| max_following | 50 | Set a limit for the number of following that will be fetched for the profile. |
| load_about_account | true | Load data from 'About this account' menu.

+3 credits. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot.

+15 mentions per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/tasks/:task_id`

---

#### Profile update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/update?load_feed_posts=true&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_feed_posts | true | Enable or disable fetching of posts from the feed section of the profile. |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the profile. |
| from_date | 2021-01-01 | Filter items by date (lower bound).

Timestamp value must be in ISO 8601 format. - Items are collected from newest to oldest and the update process will stop once the specified from_date is reached. - If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run. |
| load_feed_posts_replies | true | Enable or disable fetching of posts replies from the feed section of the profile. |
| max_feed_posts_replies | 50 | Set limit for the number of posts replies that will be fetched for the profile. |
| load_followers | true | Enable or disable fetching of followers of the profile.

+3 credit per follower. |
| max_followers | 50 | Set a limit for the number of followers that will be fetched for the profile. |
| load_following | true | Enable or disable fetching of following of the profile.

+3 credit per following. |
| max_following | 50 | Set a limit for the number of following that will be fetched for the profile. |
| load_about_account | true | Load data from 'About this account' menu.

+3 credits. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.

+1 mention per profile. |
| page_screenshot | true | Take full page screenshot.

+15 mentions per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the profile task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/update`

---

#### Cached profile data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id`

---

#### Cached profile posts

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/feed/posts?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2021-01-01 | Filter items by date (lower bound).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| lang | en | Filter items by text language.

Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| location_name | ("New York" or NY) -park | Filter items by the location name.

A query can contain quotes, logical operators, parenthesis and stopwords. |
| query | united states | Filter items by the text search query.

A query can contain quotes, logical operator OR and stopwords.
Query language can be chosen with the parameter query_lang. This will increase the filtering quality. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached profile followers profiles

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/followers/profiles?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | aWRfZGVzY3wxMzQ1MDM1OTY3 | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached profile following profiles

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/following/profiles?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

### search

---

#### for posts

---

##### auto-update task

---

###### Post search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/twitter/search/post/update?keywords=covid-19 OR (covid AND vaccine)&search_type=latest&max_posts=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | covid-19 OR (covid AND vaccine) | List of keywords or a phrase to search

You can use logical operators (AND, OR, etc) and stopwords. |
| search_type | latest | Method to use for search

top - search for the most relevant posts
latest - search for latest posts |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the search.

We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| request | covid-19 OR (covid AND vaccine) | List of keywords or a phrase to search using advanced search

Use this parameter only if you want to write your search queries manually as it limits filtering ability of the API.
Twitter search supports logical operators (AND, OR, etc) and stopwords.
You can use Twitter Advanced search page (https://twitter.com/search-advanced) to create query. |
| from_profile | potus | Filter result by the author

Equivalent for from:{username} in Twitter search request. |
| to_profile | potus | Filter result by the reply target.

Equivalent for to:{username} in Twitter search request. |
| tagged_profile | potus | Filter result by the tagged profile.

Equivalent for @{username} in Twitter search request. |
| min_faves | 100 | Filter result by the minimal number of likes.

Equivalent for min_faves:{value} in Twitter search request. |
| min_replies | 100 | Filter result by the minimal number of replies.

Equivalent for min_replies:{value} in Twitter search request. |
| min_retweets | 100 | Filter result by the minimal number of retweets.

Equivalent for min_retweets:{value} in Twitter search request. |
| lang | en | Filter result by language.

Equivalent for lang:{lang} in Twitter search request. |
| content_filters | safe,-news | Remove some result based on content type

replies - remove all original tweets
-replies - remove all replies
links - remove all tweets without links
-links - remove all tweets with links
images - remove all tweets without images
-images - remove all tweets with images
media - remove all tweets without images or videos
-media - remove all tweets with images or videos
native_video - remove all tweets without native videos
-native_video - remove all tweets with native videos
news - remove all tweets from users
-news - remove all tweets from news agencies
safe - remove all tweets with dangerous content |
| from_date | 2021-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| load_replies | true | Enable or disable fetching of replies for the posts.

+1 mentions per post reply. |
| max_replies | 50 | Set limits for the number of replies that will be fetched for every post. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/search/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

The parameter does not affect the number of items fetched for profiles, searches, etc during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/twitter/search/post/tasks/:task_id`

---

##### Post search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/twitter/search/post/update?keywords=covid-19 OR (covid AND vaccine)&search_type=latest&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | covid-19 OR (covid AND vaccine) | List of keywords or a phrase to search

You can use logical operators (AND, OR, etc) and stopwords. |
| search_type | latest | Method to use for search

top - search for the most relevant posts
latest - search for latest posts |
| max_posts | 50 | Set limit for the number of posts that will be fetched for the search.

We do not recommend setting this parameter to values > 300 as it can cause the update process to fail. |
| from_profile | potus | Filter result by the author

Equivalent for from:{username} in Twitter search request. |
| to_profile | potus | Filter result by the reply target.

Equivalent for to:{username} in Twitter search request. |
| tagged_profile | potus | Filter result by the tagged profile.

Equivalent for @{username} in Twitter search request. |
| min_faves | 100 | Filter result by the minimal number of likes.

Equivalent for min_faves:{value} in Twitter search request. |
| min_replies | 100 | Filter result by the minimal number of replies.

Equivalent for min_replies:{value} in Twitter search request. |
| min_retweets | 100 | Filter result by the minimal number of retweets.

Equivalent for min_retweets:{value} in Twitter search request. |
| lang | en | Filter result by language.

Equivalent for lang:{lang} in Twitter search request. |
| content_filters | safe,-news | Remove some result based on content type

replies - remove all original tweets
-replies - remove all replies
links - remove all tweets without links
-links - remove all tweets with links
images - remove all tweets without images
-images - remove all tweets with images
media - remove all tweets without images or videos
-media - remove all tweets with images or videos
native_video - remove all tweets without native videos
-native_video - remove all tweets with native videos
news - remove all tweets from users
-news - remove all tweets from news agencies
safe - remove all tweets with dangerous content |
| from_date | 2021-01-01 | Filter items by date (lower bound).
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.

If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| load_replies | true | Enable or disable fetching of replies for the posts.

+1 mentions per post reply. |
| max_replies | 50 | Set limits for the number of replies that will be fetched for every post. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/search/post/update?keywords=covid-19 OR (covid AND vaccine)&search_type=latest`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | covid-19 OR (covid AND vaccine) | List of keywords or a phrase to search

You can use logical operators (AND, OR, etc) and stopwords. |
| search_type | latest | Method to use for search

top - search for the most relevant posts
latest - search for latest posts |
| from_profile | potus | Filter result by the author

Equivalent for from:{username} in Twitter search request. |
| to_profile | potus | Filter result by the reply target.

Equivalent for to:{username} in Twitter search request. |
| tagged_profile | potus | Filter result by the tagged profile.

Equivalent for @{username} in Twitter search request. |
| min_faves | 100 | Filter result by the minimal number of likes.

Equivalent for min_faves:{value} in Twitter search request. |
| min_replies | 100 | Filter result by the minimal number of replies.

Equivalent for min_replies:{value} in Twitter search request. |
| min_retweets | 100 | Filter result by the minimal number of retweets.

Equivalent for min_retweets:{value} in Twitter search request. |
| lang | en | Filter result by language.

Equivalent for lang:{lang} in Twitter search request. |
| content_filters | safe,-news | Remove some result based on content type

replies - remove all original tweets
-replies - remove all replies
links - remove all tweets without links
-links - remove all tweets with links
images - remove all tweets without images
-images - remove all tweets with images
media - remove all tweets without images or videos
-media - remove all tweets with images or videos
native_video - remove all tweets without native videos
-native_video - remove all tweets with native videos
news - remove all tweets from users
-news - remove all tweets from news agencies
safe - remove all tweets with dangerous content |
| from_date | 2021-01-01 | Filter result by date (lower bound, inclusive).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format (2007-04-06). |
| to_date | 2023-01-01 | Filter result by date (upper bound, non-inclusive).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format (2007-04-06). |

---

##### Cached posts from posts search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/search/post/posts?keywords=covid-19 OR (covid AND vaccine)&order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | covid-19 OR (covid AND vaccine) | List of keywords or a phrase to search

You can use logical operators (AND, OR, etc) and stopwords. |
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_profile | potus | Filter result by the author

Equivalent for from:{username} in Twitter search request. |
| to_profile | potus | Filter result by the reply target.

Equivalent for to:{username} in Twitter search request. |
| tagged_profile | potus | Filter result by the tagged profile.

Equivalent for @{username} in Twitter search request. |
| min_faves | 100 | Filter result by the minimal number of likes.

Equivalent for min_faves:{value} in Twitter search request. |
| min_replies | 100 | Filter result by the minimal number of replies.

Equivalent for min_replies:{value} in Twitter search request. |
| min_retweets | 100 | Filter result by the minimal number of retweets.

Equivalent for min_retweets:{value} in Twitter search request. |
| content_filters | safe,-replies | Remove some result based on content type

replies - remove all original tweets
-replies - remove all replies
links - remove all tweets without links
-links - remove all tweets with links
images - remove all tweets without images
-images - remove all tweets with images
media - remove all tweets without images or videos
-media - remove all tweets with images or videos
native_video - remove all tweets without native videos
-native_video - remove all tweets with native videos
news - remove all tweets from users
-news - remove all tweets from news agencies
safe - remove all tweets with dangerous content |
| from_date | 2021-01-01 | Filter items by date (lower bound).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2023-01-01 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| lang | en | Filter items by text language.

Language must be specified as two-letter ISO 639-1 code.
Supports multiple comma-separated values. |
| location_name | ("New York" or NY) -park | Filter items by the location name.

A query can contain quotes, logical operators, parenthesis and stopwords. |
| query | "microsoft windows" or linux -apple | Filter items by the text search query.

A query can contain quotes, logical operator OR and stopwords. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

#### for profiles

---

##### auto-update task

---

###### Profiles search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/twitter/search/profile/update?keywords=(covid OR coronavirus) -vaccine&search_type=users&max_profiles=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | (covid OR coronavirus) -vaccine | List of keywords or a phrase to search

You can use logical operators (AND, OR, etc) and stopwords. |
| search_type | users | Method to use for search

users - search for the most relevant profiles |
| max_profiles | 50 | Set limit for the number of profiles that will be fetched for the search. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| request | (covid OR coronavirus) -vaccine | List of keywords or a phrase to search using advanced search

Use this parameter only if you want to write your search queries manually as it limits filtering ability of the API.
Twitter search supports logical operators (AND, OR, etc) and stopwords.
You can use Twitter Advanced search page (https://twitter.com/search-advanced) to create query.

Example: request=(covid OR coronavirus) -vaccine since:2021-01-01 until:2021-03-01 near:"New York" |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

###### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/search/profile/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

###### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/twitter/search/profile/tasks/:task_id`

---

##### Profiles search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/twitter/search/profile/update?keywords=(covid OR coronavirus) -vaccine&search_type=users&max_profiles=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | (covid OR coronavirus) -vaccine | List of keywords or a phrase to search

You can use logical operators (AND, OR, etc) and stopwords. |
| search_type | users | Method to use for search

users - search for the most relevant profiles |
| max_profiles | 50 | Set limit for the number of profiles that will be fetched for the search. |
| load_profiles_data | true | If this option is enabled: every profile update will cost 9 mentions instead of 1. |
| upload_avatar_to_s3 | true | Upload profile avatar to external data storage.

+1 mention per profile. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/search/profile/update?keywords=(covid OR coronavirus) -vaccine&search_type=users`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | (covid OR coronavirus) -vaccine | List of keywords or a phrase to search

You can use logical operators (AND, OR, etc) and stopwords. |
| search_type | users | Method to use for search

users - search for the most relevant profiles |

---

##### Cached profiles from profiles search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/search/profile/profiles?keywords=(covid OR coronavirus) -vaccine&max_page_size=50&order_by=id_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | (covid OR coronavirus) -vaccine | List of keywords or a phrase to search

You can use logical operators (AND, OR, etc) and stopwords. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

### post

---

#### auto-update task

---

##### Post auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/feed/posts/:post_id/update?load_replies=true&max_replies=50&sort_replies_by=relevancy&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_replies | true | Enable or disable fetching of replies for every post. |
| max_replies | 50 | Set limits for the number of replies that will be fetched for every post. |
| sort_replies_by | relevancy | Defines a field to order post replies by and sort direction.

Available options are:

relevancy (default): Sorts replies based on relevance scoring.
latest: Sorts replies from newest to oldest.
likes: Sorts replies by the number of likes in descending order.
+3 credits are charged only for posts that have replies and only if the sorting method is set to either latest or likes. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| load_quotes | true | Enable or disable fetching of quotes for the post.

+1 mentions per post quote. |
| max_quotes | 50 | Set limits for the number of replies that will be fetched for the post. |
| sort_quotes_by | top | Specifies the field and direction used to sort post quotes.

Available options are:

top (default): Sorts by Twitter's internal ranking of top quotes.
recent: Sorts from newest to oldest.
+2 credits are charged only for posts that have quotes and only if the sorting method is set to recent. |
| load_retweets_profiles | true | Enable or disable fetching of retweets profiles for post.

+1 mentions per profile. |
| max_retweets_profiles | 50 | Set limits for the number of retweets profiles that will be fetched for post. |
| sort_retweets_by | top | Specifies the field and direction used to sort retweets.

Available options are:

top (default): Sorts by Twitter's internal ranking of top retweets.
recent: Sorts from newest to oldest.
+2 credits are charged only for posts that have retweets and only if the sorting method is set to recent. |
| from_retweet_id | 123456789 | Set a retweet ID to start fetching retweets from. If not set, the task will fetch all retweets. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/twitter/post/tasks/:task_id`

---

#### Post update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/feed/posts/:post_id/update?load_replies=true&max_replies=50&sort_replies_by=relevancy`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| load_replies | true | Enable or disable fetching of replies for every post. |
| max_replies | 50 | Set limits for the number of replies that will be fetched for every post. |
| sort_replies_by | relevancy | Defines a field to order post replies by and sort direction.

Available options are:

relevancy (default): Sorts replies based on relevance scoring.
latest: Sorts replies from newest to oldest.
likes: Sorts replies by the number of likes in descending order.
+3 credits are charged only for posts that have replies and only if the sorting method is set to either latest or likes. |
| load_quotes | true | Enable or disable fetching of quotes for the post.

+1 mentions per post quote. |
| max_quotes | 50 | Set limits for the number of replies that will be fetched for the post. |
| sort_quotes_by | top | Specifies the field and direction used to sort post quotes.

Available options are:

top (default): Sorts by Twitter's internal ranking of top quotes.
recent: Sorts from newest to oldest.
+2 credits are charged only for posts that have quotes and only if the sorting method is set to recent. |
| load_retweets_profiles | true | Enable or disable fetching of retweets profiles for post.

+1 mentions per profile. |
| max_retweets_profiles | 50 | Set limits for the number of retweets profiles that will be fetched for post. |
| sort_retweets_by | top | Specifies the field and direction used to sort retweets.

Available options are:

top (default): Sorts by Twitter's internal ranking of top retweets.
recent: Sorts from newest to oldest.
+2 credits are charged only for posts that have retweets and only if the sorting method is set to recent. |
| from_retweet_id | 123456789 | Set a retweet ID to start fetching retweets from. If not set, the task will fetch all retweets. |
| callback_url | https://webhook.site/d63fd0b0-efb3-4138-9aff-3525f4a3f0fb | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the post task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/feed/posts/:post_id/update`

---

#### Cached post data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/feed/posts/:post_id`

---

#### Cached post replies

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/feed/posts/:post_id/replies?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | aWRfYXNjfDE2NDI3NzI2NDUyMjA0MzgwMTY= | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached profiles for post retweets

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/feed/posts/:post_id/retweets?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction.Sorting by id_desc allows to get the order of retweets on Twitter |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | aWRfZGVzY3wxMzExNDI4OTg= | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

#### Cached post quotes

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/profile/:profile_id/feed/posts/:post_id/quotes?order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | aWRfZGVzY3wxNjY2MTcyOTMzMDE2ODk5NTg0 | Pagination cursor for the next page.
Can be retrieved from the previous page (data.page_info.cursor). |

---

### Twitter queue size

**GET** `{{API_BASE_URL}}{{API_VERSION}}/twitter/queues/size`

---

### API usage stats

**GET** `{{API_BASE_URL}}{{API_VERSION}}/stats/usage?from_date=2025-03-25&to_date=2025-04-01&services=facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| from_date | 2025-03-25 | Calculate usage from this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| to_date | 2025-04-01 | Calculate usage to this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| services | facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest | Filter usage stats by services.

Multiple values can be provided as a comma-separated list. |

---

## reddit

---

### post

---

#### auto-update task

---

##### Post auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/reddit/post/:post_id/update?max_comments=50&depth=2&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_comments | 50 | Enable (if > 0) or disable (if = 0) fetching of comments for the post.

+1 credit per 10 comments, minimum +4 credits |
| depth | 2 | Defines how many reply levels are fetched for each comment.

Use depth = 0 to fetch all levels without limit.
depth = 1 to fetch only top-level comments.
depth = 2 to fetch comments +1st level replies.
depth = 3 to fetch comments +1st +2nd level replies, and so on. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| comments_sort_type | new | Defines how comments are sorted.

best
top
new
old
controversial
random
qa
live |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/reddit/post/tasks/:task_id`

---

#### Post update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/reddit/post/:post_id/update?max_comments=50&depth=2`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_comments | 50 | Enable (if > 0) or disable (if = 0) fetching of comments for the post.

+1 credit per 10 comments, minimum +4 credits |
| depth | 2 | Defines how many reply levels are fetched for each comment.

Use depth = 0 to fetch all levels without limit.
depth = 1 to fetch only top-level comments.
depth = 2 to fetch comments +1st level replies.
depth = 3 to fetch comments +1st +2nd level replies, and so on. |
| comments_sort_type | new | Defines how comments are sorted.

best
top
new
old
controversial
random
qa
live |
| callback_url | https://webhook.site/d63fd0b0-efb3-4138-9aff-3525f4a3f0fb | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the post task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/post/:post_id/update`

---

#### Cached post data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/post/:post_id`

---

#### Cached post comments

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/post/:post_id/comments?order_by=date_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| order_by | date_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2025-08-01 | Filter items by date (lower bound).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2025-08-28 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | MjY3NDI2NjQ0ODMxNzk4 | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

### comment

---

#### auto-update task

---

##### Comment  auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/reddit/comment/:comment_id/update?auto_update_interval=300&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| auto_update_interval | 300 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/comment/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/reddit/comment/tasks/:task_id`

---

#### Comment update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/reddit/comment/:comment_id/update`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| callback_url | https://webhook.site/d63fd0b0-efb3-4138-9aff-3525f4a3f0fb | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the comment task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/comment/:comment_id/update`

---

#### Cached comment data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/comment/:comment_id`

---

### post search

---

#### auto-update task

---

##### Post search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/post/update?keywords=tesla&max_posts=50&sort_type=new&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| max_posts | 50 | Maximum number of posts to fetch.

+1 credit per post. |
| sort_type | new | Method to use for search

relevance
hot
top
new
comments - search for posts with the most comments |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| date_posted | null | Search time limit. Date_posted is allowed only with sort_type values: comments, top, relevance.

all_time
past_year
past_month
past_week
today
past_hour |
| from_date | 2026-01-01 | Filter items by date (lower bound). From_date is allowed only with sort_type value: new.

Timestamp value must be in ISO 8601 format.
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.
If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run. |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |
| max_comments | 50 | Enable (if > 0) or disable (if = 0) fetching of comments for the post.

+1 credit per 10 comments, minimum +4 credits |
| depth | 2 | Defines how many reply levels are fetched for each comment.

Use depth = 0 to fetch all levels without limit.
depth = 1 to fetch only top-level comments.
depth = 2 to fetch comments +1st level replies.
depth = 3 to fetch comments +1st +2nd level replies, and so on. |
| comments_sort_type | new | Defines how comments are sorted.

best
top
new
old
controversial
random
qa
live |
| callback_url | https://webhook.site/da51754b-e7a1-4bb1-87e1-e317a8cee5f3 | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/post/tasks/:task_id`

---

#### Post search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/post/update?keywords=tesla&max_posts=50&sort_type=new`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| max_posts | 50 | Maximum number of posts to fetch.

+1 credit per post. |
| sort_type | new | Method to use for search

relevance
hot
top
new
comments - search for posts with the most comments |
| date_posted | null | Search time limit. Date_posted is allowed only with sort_type values: comments, top, relevance.

all_time
past_year
past_month
past_week
today
past_hour |
| from_date | 2026-01-01 | Filter items by date (lower bound). From_date is allowed only with sort_type value: new.

Timestamp value must be in ISO 8601 format.
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.
If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run. |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |
| max_comments | 50 | Enable (if > 0) or disable (if = 0) fetching of comments for the post.

+1 credit per 10 comments, minimum +4 credits |
| depth | 2 | Defines how many reply levels are fetched for each comment.

Use depth = 0 to fetch all levels without limit.
depth = 1 to fetch only top-level comments.
depth = 2 to fetch comments +1st level replies.
depth = 3 to fetch comments +1st +2nd level replies, and so on. |
| comments_sort_type | new | Defines how comments are sorted.

best
top
new
old
controversial
random
qa
live |
| callback_url | https://webhook.site/d63fd0b0-efb3-4138-9aff-3525f4a3f0fb | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the search post task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/post/update?keywords=tesla&sort_type=new&date_posted=all_time`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| sort_type | new | Method to use for search

relevance
hot
top
new
comments - search for posts with the most comments |
| date_posted | all_time | Search time limit. Date_posted is allowed only with sort_type values: comments, top, relevance.

all_time
past_year
past_month
past_week
today
past_hour |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |

---

#### Post search task details

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/post?keywords=tesla&sort_type=new&date_posted=all_time`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| sort_type | new | Method to use for search

relevance
hot
top
new
comments - search for posts with the most comments |
| date_posted | all_time | Search time limit. Date_posted is allowed only with sort_type values: comments, top, relevance.

all_time
past_year
past_month
past_week
today
past_hour |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |

---

#### Cached posts from search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/post/items?keywords=tesla&sort_type=new&date_posted=all_time&max_page_size=50&order_by=id_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| sort_type | new | Method to use for search

relevance
hot
top
new
comments - search for posts with the most comments |
| date_posted | all_time | Search time limit. Date_posted is allowed only with sort_type values: comments, top, relevance.

all_time
past_year
past_month
past_week
today
past_hour |
| max_page_size | 50 | Defines the maximal number of items on every results page.

The parameter does not affect the number of items fetched for profiles, searches, etc during the updating process. |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |
| from_date | 2025-08-01 | Filter items by date (lower bound).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2025-08-28 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | aWRfZGVzY3wzMjQyMDM3OTUx | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

### profile search

---

#### auto-update task

---

##### Profile search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/profile/update?keywords=tesla&max_profiles=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| max_profiles | 50 | Maximum number of posts to fetch.

+1 credit per post. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |
| callback_url | https://webhook.site/da51754b-e7a1-4bb1-87e1-e317a8cee5f3 | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/profile/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/profile/tasks/:task_id`

---

#### Task to update profile search

**POST** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/profile/update?keywords=tesla&max_profiles=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| max_profiles | 50 | Maximum number of profiles to fetch. Reddit profile search returns no more than 100 profiles now.

+1 credit per profile. |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |
| callback_url | https://webhook.site/d63fd0b0-efb3-4138-9aff-3525f4a3f0fb | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/profile/update?keywords=tesla`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |

---

#### Profile search task details

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/profile?keywords=tesla`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |

---

#### Cached profiles from search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/profile/items?keywords=tesla&max_page_size=50&order_by=id_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| max_page_size | 50 | Defines the maximal number of items on every results page.

The parameter does not affect the number of items fetched for profiles, searches, etc during the updating process. |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |
| from_date | 2018-05-01 | Filter items by date (lower bound).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2018-05-31 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | aWRfZGVzY3wzMjQyMDM3OTUx | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

### subreddit search

---

#### auto-update task

---

##### Subreddit search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/subreddit/update?keywords=tesla&max_subreddits=50&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| max_subreddits | 50 | Maximum number of subreddits to fetch. Reddit subreddit search returns no more than 250 subreddits now.

+1 credit per subreddit. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |
| callback_url | https://webhook.site/da51754b-e7a1-4bb1-87e1-e317a8cee5f3 | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/subreddit/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/subreddit/tasks/:task_id`

---

#### Task to update subreddit search

**POST** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/subreddit/update?keywords=tesla&max_subreddits=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| max_subreddits | 50 | Maximum number of subreddits to fetch. Reddit subreddit search returns no more than 250 subreddits now.

+1 credit per subreddit. |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |
| callback_url | https://webhook.site/d63fd0b0-efb3-4138-9aff-3525f4a3f0fb | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the search task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/subreddit/update?keywords=tesla`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |

---

#### Subreddit search task details

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/subreddit?keywords=tesla`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |

---

#### Cached subreddits from search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/search/subreddit/items?keywords=tesla&max_page_size=50&order_by=id_desc`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | tesla | List of keywords or a phrase to search |
| max_page_size | 50 | Defines the maximal number of items on every results page.

The parameter does not affect the number of items fetched for profiles, searches, etc during the updating process. |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| include_over_18 | true | Enable or disable search with NSFW content. Search will include NSFW content if "include_over_18" is true.

+20 credits per search. |
| from_date | 2018-05-01 | Filter items by date (lower bound).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2018-05-31 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | aWRfZGVzY3wzMjQyMDM3OTUx | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

### subreddit

---

#### auto-update task

---

##### Subreddit auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/reddit/subreddit/:name/update?max_posts=50&sort_type=new&auto_update_interval=86400&auto_update_expire_at=2027-01-01T00:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_posts | 50 | Maximum number of posts to fetch.

+1 credit per post. |
| sort_type | new | Method to use for search

best
hot
top
new
rising
controversial |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T00:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| date_posted | all_time | Search time limit. Date_posted is allowed only with sort_type values: top, controversial.

all_time
past_year
past_month
past_week
today
past_hour |
| from_date | 2026-01-01 | Filter items by date (lower bound). From_date is allowed only with sort_type value: new.

Timestamp value must be in ISO 8601 format.
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.
If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run. |
| max_comments | 50 | Enable (if > 0) or disable (if = 0) fetching of comments for the post.

+1 credit per 10 comments, minimum +4 credits |
| depth | 2 | Defines how many reply levels are fetched for each comment.

Use depth = 0 to fetch all levels without limit.
depth = 1 to fetch only top-level comments.
depth = 2 to fetch comments +1st level replies.
depth = 3 to fetch comments +1st +2nd level replies, and so on. |
| comments_sort_type | new | Defines how comments are sorted.

best
top
new
old
controversial
random
qa
live |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/subreddit/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/reddit/subreddit/tasks/:task_id`

---

#### Subreddit update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/reddit/subreddit/:name/update?max_posts=50&sort_type=new`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| max_posts | 50 | Maximum number of posts to fetch.

+1 credit per post. |
| sort_type | new | Method to use for search

best
hot
top
new
rising
controversial |
| date_posted | all_time | Search time limit. Date_posted is allowed only with sort_type values: top, controversial.

all_time
past_year
past_month
past_week
today
past_hour |
| from_date | 2026-01-01 | Filter items by date (lower bound). From_date is allowed only with sort_type value: new.

Timestamp value must be in ISO 8601 format.
Items are collected from newest to oldest and the update process will stop once the specified from_date is reached.
If used in an auto update task, the from_date value will automatically shift forward by the length of the update interval for each subsequent run. |
| max_comments | 50 | Enable (if > 0) or disable (if = 0) fetching of comments for the post.

+1 credit per 10 comments, minimum +4 credits |
| depth | 2 | Defines how many reply levels are fetched for each comment.

Use depth = 0 to fetch all levels without limit.
depth = 1 to fetch only top-level comments.
depth = 2 to fetch comments +1st level replies.
depth = 3 to fetch comments +1st +2nd level replies, and so on. |
| comments_sort_type | new | Defines how comments are sorted.

best
top
new
old
controversial
random
qa
live |
| callback_url | https://webhook.site/d63fd0b0-efb3-4138-9aff-3525f4a3f0fb | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the subreddit task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/subreddit/:name/update`

---

#### Subreddit task details

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/subreddit/:name`

---

#### Cached posts from subreddit

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/subreddit/:name/items?sort_type=new&order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| sort_type | new | Method to use for search

best
hot
top
new
rising
controversial |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2025-04-06 | Filter items by date (lower bound).

All items created before this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| to_date | 2025-05-06 | Filter items by date (upper bound).

All items created after and at this date will be removed from the response.
Timestamp value must be in ISO 8601 format. |
| cursor | MjY3NDI2NjQ0ODMxNzk4 | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

### Reddit queue size

**GET** `{{API_BASE_URL}}{{API_VERSION}}/reddit/queues/size`

---

### API usage stats

**GET** `{{API_BASE_URL}}{{API_VERSION}}/stats/usage?from_date=2025-03-25&to_date=2025-04-01&services=facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| from_date | 2025-03-25 | Calculate usage from this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| to_date | 2025-04-01 | Calculate usage to this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| services | facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest | Filter usage stats by services.

Multiple values can be provided as a comma-separated list. |

---

## threads

---

### post

---

#### auto-update task

---

##### Post auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/threads/post/:post_id/update?auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/threads/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page. [1 .. 100] |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/threads/post/tasks/:task_id`

---

#### Post update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/threads/post/:post_id/update`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| callback_url | https://webhook.site/d63fd0b0-efb3-4138-9aff-3525f4a3f0fb | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the post task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/threads/post/:post_id/update`

---

#### Cached post data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/threads/post/:post_id`

---

### post search

---

#### auto-update task

---

##### Post search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/threads/search/post/update?keywords=grey&max_posts=50&sort_type=recent&auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | grey | List of keywords or a phrase to search |
| max_posts | 50 | Maximum number of posts to fetch.

+1 credit per post. |
| sort_type | recent | Method to use for search

relevance
hot
top
new
comments - search for posts with the most comments |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| author_username | cryptorevelation | author's username |
| from_date | 2025-12-01 | Filter items by date (lower bound).

Timestamp value must be in ISO 8601 format. |
| to_date | 2025-12-31 | Filter items by date (upper bound).

Timestamp value must be in ISO 8601 format. |
| callback_url | https://webhook.site/da51754b-e7a1-4bb1-87e1-e317a8cee5f3 | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/threads/search/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/threads/search/post/tasks/:task_id`

---

#### Post search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/threads/search/post/update?keywords=grey&max_posts=50&sort_type=recent`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | grey | List of keywords or a phrase to search |
| max_posts | 50 | Maximum number of posts to fetch.

+1 credit per post. |
| sort_type | recent | Method to use for search

relevance
hot
top
new
comments - search for posts with the most comments |
| author_username | cryptorevelation | author's username |
| from_date | 2025-12-12 | Filter items by date (lower bound).

Timestamp value must be in ISO 8601 format. |
| to_date | 2025-12-02 | Filter items by date (upper bound).

Timestamp value must be in ISO 8601 format. |
| callback_url | https://webhook.site/4e4b9551-bdcb-4453-8ecb-f636a3859f86 | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the search post task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/threads/search/post/update?keywords=grey&max_posts=50&sort_type=recent`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | grey | List of keywords or a phrase to search |
| max_posts | 50 | Maximum number of posts to fetch.

+1 credit per post. |
| sort_type | recent | Method to use for search

relevance
hot
top
new
comments - search for posts with the most comments |
| author_username | cryptorevelation | author's username |
| from_date | 2025-12-01 | Filter items by date (lower bound).

Timestamp value must be in ISO 8601 format. |
| to_date | 2025-12-31 | Filter items by date (upper bound).

Timestamp value must be in ISO 8601 format. |

---

#### Post search task details

**GET** `{{API_BASE_URL}}{{API_VERSION}}/threads/search/post?keywords=grey&sort_type=recent`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | grey | List of keywords or a phrase to search |
| sort_type | recent | Method to use for search

relevance
hot
top
new
comments - search for posts with the most comments |
| author_username | cryptorevelation | author's username |
| from_date | 2025-12-01 | Filter items by date (lower bound).

Timestamp value must be in ISO 8601 format. |
| to_date | 2025-12-31 | Filter items by date (upper bound).

Timestamp value must be in ISO 8601 format. |

---

#### Cached posts from search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/threads/search/post/items?keywords=grey&sort_type=recent&order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | grey | List of keywords or a phrase to search |
| sort_type | recent | Method to use for search

relevance
hot
top
new
comments - search for posts with the most comments |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| author_username | cryptorevelation | author's username |
| from_date | 2025-12-12 | Filter items by date (lower bound).

Timestamp value must be in ISO 8601 format. |
| to_date | 2025-12-02 | Filter items by date (upper bound).

Timestamp value must be in ISO 8601 format. |
| cursor | MjY3NDI2NjQ0ODMxNzk4 | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

### Threads queue size

**GET** `{{API_BASE_URL}}{{API_VERSION}}/threads/queues/size`

---

### API usage stats

**GET** `{{API_BASE_URL}}{{API_VERSION}}/stats/usage?from_date=2025-03-25&to_date=2025-04-01&services=facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| from_date | 2025-03-25 | Calculate usage from this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| to_date | 2025-04-01 | Calculate usage to this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| services | facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest | Filter usage stats by services.

Multiple values can be provided as a comma-separated list. |

---

## pinterest

---

### post search

---

#### auto-update task

---

##### Post search auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/search/post/update?keywords=grey&max_posts=50&auto_update_interval=86400&auto_update_expire_at=2026-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | grey | List of keywords or a phrase to search |
| max_posts | 50 | Maximum number of posts to fetch.

+1 credit per post. |
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2026-01-01T15:00:00 | Auto update task will be canceled after this date.

Timestamp value must be in ISO 8601 format (2007-04-06T21:42). |
| callback_url | https://webhook.site/da51754b-e7a1-4bb1-87e1-e317a8cee5f3 | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/search/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/search/post/tasks/:task_id`

---

#### Post search update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/search/post/update?keywords=grey&max_posts=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | grey | List of keywords or a phrase to search |
| max_posts | 50 | Maximum number of posts to fetch.

+1 credit per post. |
| callback_url | https://webhook.site/4e4b9551-bdcb-4453-8ecb-f636a3859f86 | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the search post task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/search/post/update?keywords=grey`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | grey |  |

---

#### Post search task details

**GET** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/search/post?keywords=grey`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | grey | List of keywords or a phrase to search |

---

#### Cached posts from search

**GET** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/search/post/items?keywords=grey&order_by=id_desc&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| keywords | grey | List of keywords or a phrase to search |
| order_by | id_desc | Defines a field to order items by and sort direction. |
| max_page_size | 50 | Defines the maximal number of items on every results page.

If the value is greater than maximum, the system will automatically set it to maximum.
The parameter does not affect the number of items fetched during the updating process. |
| from_date | 2025-12-12 | Filter items by date (lower bound).

Timestamp value must be in ISO 8601 format. |
| to_date | 2025-12-02 | Filter items by date (upper bound).

Timestamp value must be in ISO 8601 format. |
| cursor | MjY3NDI2NjQ0ODMxNzk4 | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

### post

---

#### auto-update task

---

##### Post auto-update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/post/:post_id/update?auto_update_interval=86400&auto_update_expire_at=2027-01-01T15:00:00`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| auto_update_interval | 86400 | Create an auto update task. Update task for the item will be issued every auto_update_interval seconds.

API will make a POST request to the callback URL every time the update process has finished.
This parameter is useful if you want to need to monitor items, but don't want to create update tasks manually.
You can retrieve or cancel auto update tasks with /tasks endpoints. |
| auto_update_expire_at | 2027-01-01T15:00:00 | Auto update task will be canceled after this date.

The timestamp must be provided either in ISO 8601 format or as a negative integer value.
If the value is ISO 8601 (2007-04-06T21:42), it is treated as an absolute point in time.
If the value is a negative integer, it is interpreted as a duration in seconds subtracted from the current UTC time. |
| callback_url | https://webhook.site/0d69a2b3-bab0-41aa-8f4c-bb4f6889616b | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

##### List of the created tasks

**GET** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/post/tasks?task_type=auto_update&max_page_size=50`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| task_type | auto_update | Select what tasks to retrieve.

"update" - simple queued task for updating an item once.
"auto_update" - tasks for monitoring items for some time. |
| max_page_size | 50 | Defines the maximal number of items on every results page. [1 .. 100] |
| cursor | MTYyNjQ0OTA1MA== | Pagination cursor for the next page.

Can be retrieved from the previous page (data.page_info.cursor). |

---

##### Cancel created task

**DELETE** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/post/tasks/:task_id`

---

#### Post update task

**POST** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/post/:post_id/update`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| callback_url | https://webhook.site/fb1ae462-330d-4d88-a141-83fbfbed9319 | Webhook URL to receive loaded data.

API will make a POST request to this URL when the update process finished.
The request body will contain status and error info along with loaded data and pagination cursors.
This parameter is useful if you want to avoid API polling while waiting for updated data. |

---

#### Status of the post task

**GET** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/post/:post_id/update`

---

#### Cached post data

**GET** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/post/:post_id`

---

### Pinterest queue size

**GET** `{{API_BASE_URL}}{{API_VERSION}}/pinterest/queues/size`

---

### API usage stats

**GET** `{{API_BASE_URL}}{{API_VERSION}}/stats/usage?from_date=2025-03-25&to_date=2025-04-01&services=facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| from_date | 2025-03-25 | Calculate usage from this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| to_date | 2025-04-01 | Calculate usage to this date.

Timestamp value must be in ISO 8601 format (2007-04-06). |
| services | facebook,instagram,linkedin,tiktok,twitter,reddit,threads,pinterest | Filter usage stats by services.

Multiple values can be provided as a comma-separated list. |

---

## authorization

---

### Create an expiration subtoken

**GET** `{{API_BASE_URL}}{{API_VERSION}}/auth/access_token?expire_at=2023-01-01`

**Query Parameters:**

| Key | Value | Description |
|-----|-------|-------------|
| expire_at | 2023-01-01 | Possibility to create a token with a limited expiration date.

The consumption of the mentions for the main token and for the created tokens will be summed up. |

---
