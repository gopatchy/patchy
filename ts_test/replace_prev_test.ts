import * as test from './test.js';

test.def('replace prev success', async (t: test.T) => {
	const create = await t.client.createTestType({text: 'foo', num: 5});
	const get1 = await t.client.getTestType(create.id);
	await t.client.replaceTestType(create.id, {text: 'bar'});

	t.rejects(t.client.replaceTestType(create.id, {text: 'zig'}, {prev: get1}));

	const get2 = await t.client.getTestType(create.id);
	t.equal(get2.text, 'bar');
});
