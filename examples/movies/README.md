# Movie Agent
- Run:
  - `export TMDB_KEY=<your tmdb API key> OPENAI_KEY=<your openai key>`
  - `go run .`
- This agent can use three tools:
  - Get the current date/time
  - Search for movies based on keywords
  - Search for movies based on name
- Example conversation:
```
🤖 Chat started — type 'exit' or 'quit' to end.

You: How long ago was the original how to train your dragon released, and what was it rated?
        To answer the question, I need to determine today's date and find information about the original ...
        $ tool://get_time?
        > Thu Oct  9 09:26:51 2025
        $ tool://movie_lookup?query=how%20to%20train%20your%20dragon
        > How to Train Your Dragon (2025-06-06) — Rating 8.0/10 On the rugged isle of Berk, where Vikings...
        The current date is October 9, 2025. The original 'How to Train Your Dragon' was released on Marc...
Agent: The original 'How to Train Your Dragon' was released on March 18, 2010 and was rated 7.9/10. As of today (October 9, 2025), it was released 15 years, 6 months, and 21 days ago.

You: What abot the second one?
        The second main entry in the movie series, 'How to Train Your Dragon 2', was released on June 5, ...
Agent: 'How to Train Your Dragon 2' was released on June 5, 2014, and was rated 7.7/10. As of today (October 9, 2025), it was released 11 years, 4 months, and 4 days ago.
```