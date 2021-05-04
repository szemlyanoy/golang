Program processes text data passed in the form of file and outputs json with at most 100 most popular 3-word phrases sorted in descend order.  For performance speed up concurency was implemented to process each paragraph in separate goruotine. Results are then aggregated and processed additionaly to deduplicate results.

v1-- Sergii Zemlianyi | 03/30/2021

v2-- Sergii Zemlianyi | 05/04/2021
   + optimized final aggregation using maps      
   + split files


TODO:
-- to add unit tests 
-- concurency for final deduplication


Steps to test program
1. Copy 'phrases_popularity' binary to any Linux host
2. Run 
 ./phrases_popularity <file1>...<fileN>

where fileN - any file to process to find phrases popularity. Good sample source http://www.gutenberg.org/cache/epub/2009/pg2009.txt 
