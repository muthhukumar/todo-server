package data

import (
	"math/rand"
)

var Quotes = []string{
	"If you want to do something, don't hesitate unless it's a bad thing.",
	"Set Your Heart Ablaze",
	"Greatness from small beginnings",
	"Don’t waste time on doing things that do not provide value to our life. Sometime wasting time is fine but all we do is waste time then that is not good.",
	"When faced with a problem if you have multiple solution but unable to decide which to use just pick one that is better for that situation and just use it. Don’t struct doing nothing. Even if that solution does not provide multiple value it is fine.",
	"When faced with any problem If you have a solution that is not perfect but that can fix even 1% of that problem then do it. It is better than doing nothing.",
	"Just don’t die with regrets.",
	"We are all going to die one day, so nothing you do matters. It's not a big deal, so just chill.",
	"Absolutely do some physical activity daily. Even force yourself to do it. Maybe just do skipping.",
	"Do Skipping in the morning, Then cold shower, No Coffee, Less screen time, Do something you like, Social more, Learning something new, Meditate, On Stress Take a 3 long breath",
	"Don't be scared to do things. Push through it. It's better to worry about the aftermath than worrying about why it might not work. It's better to do it than not to do it. ",
	"There are cities filled with tall buildings, and we can visit those places. But instead, we're wasting our time doing nothing. There is so much to explore in this world, and we can explore those things. All we have to do is put in more effort, and we can achieve anything. Anything can be achieved, but only if we truly want it. And I want to want it.",
	"We're waiting for the world to tell us what to do, and it's not good. If we want to do anything, we have to act on our own without waiting for its intervention. We're waiting for the perfect opportunity or the perfect time to do anything, or we're waiting for something bad to happen and then reacting to it. It's like we're being reactive instead of proactive.",
	"Don't buy anything based on gut feeling. As they literally want you to buy based on that feeling rather than how the product is.",
	"If you play to not lose, you'll lose. But if you play to win, you might win.",
	"We should not be selfish to ourselves.",
	"Don't be scared, just do it, okay? You can do it. Even if it takes time, do it slowly, but make sure to do it.",
	"Leave things for our future self better by doing things that need to be done today. That is all you have to do.",
	"When faced with a problem you can't solve, instead of keep struggling and not finding any solution for sometime, why not ask for help from others?",
	"Preserve dopamine. Refrain from useless activities that gives dopamine hit but not productive. Like scrolling through shorts.",
	"Sugar is the new cigarette. Atleast for us.",
	"Make good decisions. They have a compounding effect. One good decision leads to another.",
	"Ask questions. Why?",
	"Sugar is poisonous.",
	"Our actions are influenced by dopamine. When we engage in activities we enjoy, we receive a surge of dopamine. However, high-dopamine activities cause a spike on the dopamine scale that does not last long. When this spike is very high, it leads to a corresponding crash afterward.",
	"Take care of your body and mind as you take care of your things(laptop or something like that).",
	"When faced with a problem try to find a solution to fix it. Even if that solution isn’t perfect or the right one it is better than doing nothing. We just need to take one step at a time and fix that problem little by little. It just need to keep the thing(task) moving.",
	"The only bad workout is the workout that didn't happen.",
	"How much longer will you wait to demand the best from yourself?",
	"Causality is an influence by which one event, process, state, or object contributes to the production of another event, process, state, or object where the cause is partly responsible for the effect, and the effect is partly dependent on the cause.",
	"Circle of life - What goes around comes around!",
	"For one who has learnt to master his mind, The Mind is the greatest of friends, but for one who has failed to do so, Mind is the greatest of enemies",
	"If you want to control restless mind, focus on your breath.",
	"Take it one day at a time. Don't worry about tomorrow or the future. Just do the thing that needs to be done today properly. That's it.",
	"It was always better to just get the unpleasant things done quickly.",
	"If there's something you have to do, you can't ignore it forever. You have to get it done eventually.",
	"We can learn a lot from others, but instead of copying everything, we should pick the good stuff and make it our own. Take the helpful parts, skip the junk, and use your own ideas to make it fit like a glove.",
	"Zelanus's essence was giving his all without letting his guard down, no matter who the opponent was.",
	"Remember the saying 'garbage in, garbage out.' Its essence is that like Rimuru, who gained great power by taking in great power, what you consume or take in influences what you become.",
	"Be flexible and able to adjust to any situation, just like Zhuo Fan in the Demonic Emperor light novel.",
	"Once you start something, like a book, see it through to the end. It's okay to change your mind if it's not good, but don't give up just because you get busy. Commit to completing things you begin, whether it's a book, a task, or anything else.",
	"Remember, as we age, our bodies might not be as strong as they are now. So, appreciate your youth and all the amazing things you can do with it. Fill your days with experiences and don't look back with regret.",
	"Having good skills isn't enough. Knowing how to use them effectively is what really matters.",
	"Take some time to go over the rules, quotes, and thoughts again and again.",
	"Hone your skills and do not rely too heavily on brute strength—that's what Diablo do.",
	"For Shion, even an army of giants seemed to be nothing but food for her own improvement.",
	"Diablo, surprisingly, did not lie, and he did not say things that he could not do.",
	"The most meaningless thing to do is to do something just for the sake of doing it. Instead, think, try, fail, and try again. That is a better approach.",
	"Don't rush through a book! Take your time and enjoy the journey. Savor each word and let the story unfold naturally.",
	"If you're likely to forget something, write it down or add it to your to-do list.",
	"Rome was not built in a day.",
	"If you want to do something, you must do it, even if the world and the odds are against you.",
	"Do what you love with passion and love, even though it gives you headaches and hardships. - Dean Schneider",
	"If we don't bargain, then we are slaves.",
	"If we work hard now, putting in the effort and dedication required, we can later enjoy the things we truly love and cherish.",
	"There are lot of things to do in this world.",
	"We are not here to control others, nor are we here to be controlled by others.",
	"If you know something could cause a problem later, solve it right away. Don't let it drag on.",
	"Don't worry about what others think of you.",
	"Always keep the thing were it is taken from",
	"Do what only you can do.",
	"Ethana naal enna pannitu iruntha",
	"If something didn't work the second time, don't try again in the same way. Try to change your approach.",
	"Lessons are taught until they are learned.",
	"Never copy and paste code without understanding it.",
	"Ask questions like 'What is the problem?' and 'What is one small thing I can do to solve it?' Just focus on one thing, no matter how small, to solve the problem step by step. If something is bothering you, find a way to fix it. ",
	"Don't always assume you're right.",
	"The strongest principle of growth lies in human choice.",
	"It's challenging for me to estimate how long tasks will take.",
	"If you make a mistake and do not correct it, that is called a mistake.",
	"To doubt everything or to believe everything are two equally convenient solutions; both dispense with the necessity of reflection.",
	"People, often deceived by an illusive god, desire their own ruin.",
	"The worst enemy you can meet will always be yourself.",
	"Force without wisdom falls of its own weight.",
	"A man who cannot command himself will always be a slave.",
	"Remember to keep a clear head in difficult times.",
	"Every failure can be a step to success.",
	"It takes great talent and skill to conceal one's talent and skill.",
	"What people call fate is mostly their own stupidity.",
	"We should not be upset that others hide the truth from us when we hide it so often from ourselves.",
	"Don't forget to wear sandals when walking in the rain.",
	"We don't have to repeat ourselves. If we've said something to someone and they understand it but are not listening, we don't have to repeat ourselves unless they ask us to.",
	"Follow our rules to an extent.",
	"Follow our rules to a certain extent.",
	"One percent improvement Daily",
	"Don't postpone tasks with thoughts like 'I'll do it later' or 'maybe next time.' Procrastination only leads to more problems. The longer you wait, the worse it gets.",
	"To make progress, remember that discipline is crucial. You must consistently show up every day, no matter what happens.",
	"When you notice a problem in your life, make a plan to solve it. Work on this plan every day, monitor your progress, and adjust the plan if needed. Keep refining the plan to better address the issue.",
	"Many people are highly skilled in their fields. I aspire to achieve their level of talent. To do that, I must consistently put in hard work every day.",
	"Use the right tool for the job.",
	"There are so many things we can improve in our life.",
	"Our ego is our greatest enemy.",
	"Don't overlook the benefits of reading. ",
	"We must admit when we're wrong because acknowledging our mistakes is the first step to improvement.",
	"Do things with purpose.",
	"Complain less. It doesn't help anyone, including yourself.",
	"Question everything, just like Socrates did.",
	"We’re going to die one day.",
	"Shorter the sleep, shorter your life span.",
	"When we're doing bad thing to others we're degrading ourselves.",
	"Goal without a plan is just a wish. A plan turns wishes into achievable results.",
	"Do things properly. Avoid using shortcuts, as they only provide temporary benefits.",
	"Change begins from within yourself.",
	"If a task takes just a minute or two, go ahead and do it immediately. Completing small tasks right away can save you time and stress in the long run.",
	"Getting angry doesn't help anyone. It doesn't lead to productive outcomes.",
	"Overthinking doesn't help.",
	"Get enough sleep to boost your memory. Good sleep helps you remember things better.",
	"When making a hard decision, think about at what cost you're doing it. When facing a tough choice, consider the price you'll pay.",
	"Unhealthy sleep, Unhealthy heart",
	"No one is as interested in your material possessions as you are, so don't obsess over them. Other people don't care what you have.",
	"The best bridge between despair and hope is a good night's sleep.",
	"Doing something is better than having nothing. Even a small amount of progress is better than doing nothing at all.",
	"Don’t let your ego run wild",
	"A sleep-deprived body will cry famine in the midst of plenty",
	"Sleep boots immunity",
	"We can never forget when we try to forget.",
	"The best time to start anything was yesterday. The second best time to do it is always today.",
	"REM-sleep dreaming is information alchemy",
	"Always break big tasks into smaller bits",
	"Always prevent self mental high",
	"He who has a why to live can bear with almost any how",
	"The unexamined life is not worth living",
	"if it is not true, do not say it.",
	"If it is not right, do not do it.",
	"A thing turns into its opposite if pushed too far",
	"The only way to discover the limits of the possible is to go beyond them into the impossible",
	"Everything has it's limit. Too much of anything is not good",
	"The obstacle in the path becomes the path. ",
	"See thing for what they are. See through with the objective eye.",
	"Those who cross the obstacles thrive. Those who don't get destroyed",
	"Embrace a mindset of growth. We often learn more when we face challenges and setbacks than when we experience easy success.",
	"When you don't want to do something remember that you're feeling that way maybe because of low dopamine.",
	"Challenge yourself to be better. The biggest competition you have is with yourself.",
	"When you want to do something, concentrate on just that one thing and avoid doing many things at once. If you try to multitask, you won't be able to give your full attention and effort to any one thing.",
	"Don't ever forget that your life situation is in other people's hands, like your boss.",
	"You must avoid the idea that you can manage learning several skills at a time. Develop your powers of concentration and understand that trying to multitask will hinder your progress.",
	"It's not enough for our code to look nice; to some extent, we should look nice too.",
	"Listing out things makes them easier to see and analyze. For example, by listing out where we're lacking, we can analyze the list and start creating solutions for each problem.",
	"If we forget an old concept we learned, it's okay to stop and relearn it. No one is going to judge you. This is how we learn.",
	"Some things, like learning new things or picking up a new hobby, might not seem important right now, but they will matter eventually over time. Have faith in the dots connecting.",
	"Make sure that the work you do is both good quality and consistent. Maintaining high standards and being reliable in what you do is essential.",
	"Don't rely on just one thing. Spread your efforts and diversify. It's important to have a variety of options and not put all your eggs in one basket.",
	"Do the same things daily at the same time, like sleeping.",
	"Don't rush things. We have a lot of time. Just get better daily, little by little.",
	"People are creating great things. So we can too. We just have to put in the work",
	"Simplify things for your future self. Ensure you leave the place better than you found it, just like at a campsite.",
	"Do things even if it is uncomfortable. Sometimes, stepping out of our comfort zone and trying new things can be a bit uncomfortable, but it's essential for our growth.",
	"Life is short so take risks. You enter the world with NOTHING and you leave the world with NOTHING. You have literally NOTHING.",
	"Never argue with an idiot. They will drag you down to their level and beat you with experience.",
	"No need to stress over things you can't control.",
	"Keep working hard, like pounding a rock. It might not break after a thousand hits, but it will eventually if you keep at it.",
	"We suffer more in imagination than in reality.",
	"If we consistently overlook minor details, we might also disregard critical elements in more significant tasks.",
	"Simplify your life. Having fewer things means fewer worries.",
	"Know thyself",
	"Think of it this way: we're like computers. We read books to learn and improve our life, similar to installing software to add specific features.",
	"Set goals and take decisions to achieve that goal",
	"Instead of fixing the problem you should eliminate the problem. Find the root cause and remove the root cause",
	"The only true wisdom is in knowing you know nothing.",
	"*What made Socrates so wise was that he kept asking questions and he was always willing to debate his ideas. Life, he declared, is only worth living if you think about what you are doing. An unexamined existence is all right for cattle, but not for human beings.*",
	"Keep in mind the downside of compounding – repeating negative actions can lead to various problems.",
	"when making a decision if we are indecisive means we don’t have a specific goal in mind. If we have a specific goal in mind then we can take decisions based on a goal",
	"Believe nothing your hear and Half of what you see.",
	"Don't let others define who you are. Your self-identity should come from within. ",
	"Before doing something hurtful to someone, consider how you'd feel if they did the same to you. ",
	"Instead of wondering when your next vacation is, maybe you should setup a life you don't need to escape from.",
	"There are people experiencing worse things than us. So if you think you’re experiencing worse then you’re in for a disappointment. They would even like to have your problems.",
	"If you try hard to save time but don't use that time wisely, then there will be no use in saving it. Unless you're doing something good with it.",
	"Always remember that there is element of luck to everything.",
	"Instead of doing nothing we are not changing our future. But right now we are changing our future by learning [skill]. It will bear fruit eventually. So don't worry keep your head down and work hard on [skill]. It will not go to waste.",
	"To break the cycle of automatic behavior, we can intentionally choose to alter our actions.",
	"Watching pornography can significantly impact the brain in a negative way.",
	"If you're good at your daily job or routine, you can plan anything and execute it properly.",
	"To become an expert in a field, the key is deliberate practice: not just doing the same thing repeatedly, but challenging yourself with a task that is just beyond your current ability. Try it, analyze your performance during and after, correct any mistakes, and then repeat. Repeat this process again and again.",
	"Complete the necessary tasks first, and then you can choose to do whatever you like.",
	"Confront your fears.",
	"Good gut is important",
	"Add the task to your calendar so you don't forget.",
	"Be like prime",
	"Be like Rae",
	"Believe in yourself, perform your actions (karma), and success will follow you.",
	"If we take care of our body, we can afford a good phone.",
}

const MAX_QUOTES = 2

// TODO: pass the max quotes as the params
func GetRandomQuotes(quotes []string) []string {
	var result []string

	for len(result) < MAX_QUOTES {
		idx := rand.Intn(len(quotes))

		result = append(result, quotes[idx])
	}

	return result
}
