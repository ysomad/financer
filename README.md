# Bot

`/start` - show hello message with usage guide
`/help` - show info about all commands
`/set_currency` - sets default user currency
`{?+}{money amount} {?currency} {expense} {?category} {?date in format 20.05 or 20.05.1999}`:
    - '+', 'currency', 'category' and 'date' is optional
    - if '+' is present its earning
    - if 'currency' is present ignores default user currency and creates entry with specified one
    - if 'category' is present its searchnig for user category with similar name if there is multiple, user have to choose one after entry submit
    - if 'date' is present entry will be created for that date
`/add_category` - add new category to user
`/delete_category` - deletes user category
`/edit_category` - edit user category, only if you author of category or creates new category with new name and replaces old one

