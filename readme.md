# What is a queue?

Queue is a collection of elements, the element in the queue is first in first out.

# What is first in first out?

There is a timestamp when every element is put into queue, then you pop an element from queue, you pop the element with the lowest timestamp.
So when building a queue, we can extent this definition, when we push an element into queue, we push with a score, and when we pop an element,
we pop the element with the lowest score.

# What is a p-queue?

p-queue is a queue with many additional features.

# What features in p-queue?

These are the features we now supported:

1) A element is pushed into queue with a score.

2) When pop, we pop the element with the lowest score.

3) A element can be acknowledged, in case it is lost by client, and after a period of time, it can get by client again.
