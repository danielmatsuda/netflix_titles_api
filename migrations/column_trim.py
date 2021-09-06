import pandas as pd

# PostgreSQL's COPY FROM command can only copy ALL columns from a file.
# So, I need to manually remove unwanted columns from the original CSV.

df = pd.read_csv("C:/Users/Daniel Matsuda/Desktop/netflix_titles.csv")
df.drop(['show_id', 'cast', 'date_added', 'rating',
         'duration', 'listed_in', 'description'], inplace=True, axis=1)
df.to_csv('C:/Users/Daniel Matsuda/Desktop/trimmed_netflix_titles.csv', index=False)
